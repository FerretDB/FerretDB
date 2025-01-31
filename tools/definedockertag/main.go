// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main contains tool for defining Docker image tags on CI.
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/sethvargo/go-githubactions"
)

func main() {
	flag.Parse()

	action := githubactions.New()

	debugEnv(action)

	res, err := define(action.Getenv)
	if err != nil {
		action.Fatalf("%s", err)
	}

	setResults(action, res)
}

// result represents Docker image names and tags extracted from the environment.
type result struct {
	evaluationImages  []string
	developmentImages []string
	productionImages  []string
}

// semVerTag is a https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string,
// but with a leading `v`.
//
//nolint:lll // for readibility
var semVerTag = regexp.MustCompile(`^v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// debugEnv logs all environment variables that start with `GITHUB_` or `INPUT_`
// in debug level.
func debugEnv(action *githubactions.Action) {
	res := make([]string, 0, 30)

	for _, l := range os.Environ() {
		if strings.HasPrefix(l, "GITHUB_") || strings.HasPrefix(l, "INPUT_") {
			res = append(res, l)
		}
	}

	slices.Sort(res)

	action.Debugf("Dumping environment variables:")

	for _, l := range res {
		action.Debugf("\t%s", l)
	}
}

// Define extracts Docker image names and tags from the environment variables defined by GitHub Actions.
func define(getenv githubactions.GetenvFunc) (*result, error) {
	repo := getenv("GITHUB_REPOSITORY")

	// to support GitHub forks
	parts := strings.Split(strings.ToLower(repo), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed to split %q into owner and name", repo)
	}
	owner := parts[0]
	repo = parts[1]

	var res *result
	var err error

	switch event := getenv("GITHUB_EVENT_NAME"); event {
	case "pull_request", "pull_request_target":
		branch := strings.ToLower(getenv("GITHUB_HEAD_REF"))
		res = defineForPR(owner, repo, branch)

	case "push", "schedule", "workflow_run":
		refName := strings.ToLower(getenv("GITHUB_REF_NAME"))

		switch refType := strings.ToLower(getenv("GITHUB_REF_TYPE")); refType {
		case "branch":
			res, err = defineForBranch(owner, repo, refName)

		case "tag":
			match := semVerTag.FindStringSubmatch(refName)
			if match == nil || len(match) != semVerTag.NumSubexp()+1 {
				return nil, fmt.Errorf("unexpected git tag %q", refName)
			}

			major := match[semVerTag.SubexpIndex("major")]
			minor := match[semVerTag.SubexpIndex("minor")]
			patch := match[semVerTag.SubexpIndex("patch")]
			prerelease := match[semVerTag.SubexpIndex("prerelease")]

			var tags []string

			if prerelease == "" {
				tags = []string{
					major,
					major + "." + minor,
					major + "." + minor + "." + patch,
				}

				if major == "2" {
					tags = append(tags, "latest")
				}
			} else {
				tags = []string{major + "." + minor + "." + patch + "-" + prerelease}

				// while v2 is not GA
				if major == "2" {
					tags = append(tags, major)
				}
			}

			res = defineForTag(owner, repo, tags)

		default:
			err = fmt.Errorf("unhandled ref type %q for event %q", refType, event)
		}

	default:
		err = fmt.Errorf("unhandled event type %q", event)
	}

	if err != nil {
		return nil, err
	}

	if res == nil {
		panic("both res and err are nil")
	}

	slices.Sort(res.evaluationImages)
	slices.Sort(res.developmentImages)
	slices.Sort(res.productionImages)

	return res, nil
}

// defineForPR defines Docker image names and tags for pull requests.
func defineForPR(owner, repo, branch string) *result {
	// for branches like "dependabot/submodules/XXX"
	parts := strings.Split(branch, "/")
	branch = parts[len(parts)-1]

	res := &result{
		evaluationImages: []string{
			fmt.Sprintf("ghcr.io/%s/%s-eval:pr-%s", owner, repo, branch),
		},
		developmentImages: []string{
			fmt.Sprintf("ghcr.io/%s/%s-dev:pr-%s", owner, repo, branch),
		},
	}

	// PRs are only for testing; no Quay.io and Docker Hub repos

	return res
}

// defineForBranch defines Docker image names and tags for branch builds.
func defineForBranch(owner, repo, branch string) (*result, error) {
	// see packages.yml
	switch {
	case branch == "main":
		fallthrough
	case strings.HasPrefix(branch, "main-"):
		fallthrough
	case strings.HasPrefix(branch, "releases/"):
		branch = strings.ReplaceAll(branch, "/", "-")

	default:
		return nil, fmt.Errorf("unhandled branch %q", branch)
	}

	res := &result{
		evaluationImages: []string{
			fmt.Sprintf("ghcr.io/%s/%s-eval:%s", owner, repo, branch),
		},
		developmentImages: []string{
			fmt.Sprintf("ghcr.io/%s/%s-dev:%s", owner, repo, branch),
		},
	}

	// forks don't have Quay.io and Docker Hub orgs
	if owner != "ferretdb" {
		return res, nil
	}

	// we don't have Quay.io and Docker Hub repos for other GitHub repos
	if repo != "ferretdb" {
		return res, nil
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/4694

	// res.evaluationImages = append(res.evaluationImages, fmt.Sprintf("quay.io/ferretdb/ferretdb-eval:%s", branch))
	// res.developmentImages = append(res.developmentImages, fmt.Sprintf("quay.io/ferretdb/ferretdb-dev:%s", branch))

	// res.evaluationImages = append(res.evaluationImages, fmt.Sprintf("ferretdb/ferretdb-eval:%s", branch))
	// res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/ferretdb-dev:%s", branch))

	return res, nil
}

// defineForTag defines Docker image names and tags for prerelease tag builds.
func defineForTag(owner, repo string, tags []string) *result {
	res := new(result)

	for _, t := range tags {
		res.evaluationImages = append(res.evaluationImages, fmt.Sprintf("ghcr.io/%s/%s-eval:%s", owner, repo, t))
		res.developmentImages = append(res.developmentImages, fmt.Sprintf("ghcr.io/%s/%s-dev:%s", owner, repo, t))
		res.productionImages = append(res.productionImages, fmt.Sprintf("ghcr.io/%s/%s:%s", owner, repo, t))
	}

	// forks don't have Quay.io and Docker Hub orgs
	if owner != "ferretdb" {
		return res
	}

	// we don't have Quay.io and Docker Hub repos for other GitHub repos
	if repo != "ferretdb" {
		return res
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/4694
	// for _, t := range tags {
	// 	res.evaluationImages = append(res.evaluationImages, fmt.Sprintf("quay.io/ferretdb/ferretdb-eval:%s", t))
	// 	res.developmentImages = append(res.developmentImages, fmt.Sprintf("quay.io/ferretdb/ferretdb-dev:%s", t))
	// 	res.productionImages = append(res.productionImages, fmt.Sprintf("quay.io/ferretdb/ferretdb:%s", t))

	// 	res.evaluationImages = append(res.evaluationImages, fmt.Sprintf("ferretdb/ferretdb-eval:%s", t))
	// 	res.developmentImages = append(res.developmentImages, fmt.Sprintf("ferretdb/ferretdb-dev:%s", t))
	// 	res.productionImages = append(res.productionImages, fmt.Sprintf("ferretdb/ferretdb:%s", t))
	// }

	return res
}

// setResults sets action output parameters, summary, etc.
func setResults(action *githubactions.Action, res *result) {
	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 1, 1, 1, ' ', tabwriter.Debug)
	fmt.Fprintf(w, "\tType\tImage\t\n")
	fmt.Fprintf(w, "\t----\t-----\t\n")

	for _, image := range res.evaluationImages {
		u := imageURL(image)
		_, _ = fmt.Fprintf(w, "\tEvaluation\t[`%s`](%s)\t\n", image, u)
	}

	for _, image := range res.developmentImages {
		u := imageURL(image)
		_, _ = fmt.Fprintf(w, "\tDevelopment\t[`%s`](%s)\t\n", image, u)
	}

	for _, image := range res.productionImages {
		u := imageURL(image)
		_, _ = fmt.Fprintf(w, "\tProduction\t[`%s`](%s)\t\n", image, u)
	}

	_ = w.Flush()

	action.AddStepSummary(buf.String())
	action.Infof("%s", buf.String())

	action.SetOutput("evaluation_images", strings.Join(res.evaluationImages, ","))
	action.SetOutput("development_images", strings.Join(res.developmentImages, ","))
	action.SetOutput("production_images", strings.Join(res.productionImages, ","))
}

// imageURL returns HTML page URL for the given image name and tag.
func imageURL(name string) string {
	switch {
	case strings.HasPrefix(name, "ghcr.io/"):
		return fmt.Sprintf("https://%s", name)
	case strings.HasPrefix(name, "quay.io/"):
		return fmt.Sprintf("https://%s", name)
	}

	name, _, _ = strings.Cut(name, ":")

	// there is not easy way to get Docker Hub URL for the given tag
	return fmt.Sprintf("https://hub.docker.com/r/%s/tags", name)
}
