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

// Package main contains script that generates changes for the latest version.
package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"
	"text/template"

	"github.com/FerretDB/gh"
	"github.com/google/go-github/v70/github"
)

//go:embed changelog.md.template
var templateFile string

// PRData represents template information about a single pull request.
type PRData struct {
	Title  string
	Author string
	URL    string
}

// Data represents template information about the release.
type Data struct {
	PrevVersion   string
	Version       string
	Date          string
	Milestone     int
	NewFeatures   []PRData
	FixedBugs     []PRData
	Enhancements  []PRData
	Documentation []PRData
	OtherChanges  []PRData
}

// getMilestone returns milestone by title (which matches FerretDB version and Git tag).
func getMilestone(ctx context.Context, client *github.Client, title string) (*github.Milestone, error) {
	opts := &github.MilestoneListOptions{
		State:     "all",
		Sort:      "due_on",
		Direction: "desc",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		milestones, resp, err := client.Issues.ListMilestones(ctx, "FerretDB", "FerretDB", opts)
		if err != nil {
			return nil, err
		}

		for _, milestone := range milestones {
			if *milestone.Title == title {
				return milestone, nil
			}
		}

		if resp.NextPage == 0 {
			return nil, fmt.Errorf("no milestone found with the title %q", title)
		}

		opts.ListOptions.Page = resp.NextPage
	}
}

// getPRs returns all pull requests for the given milestone.
func getPRs(ctx context.Context, client *github.Client, milestone *github.Milestone) ([]*github.Issue, error) {
	opts := &github.IssueListByRepoOptions{
		Milestone: strconv.Itoa(*milestone.Number),
		State:     "all",
		Sort:      "created",
		Direction: "asc",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	var prs []*github.Issue

	for {
		issues, resp, err := client.Issues.ListByRepo(ctx, "FerretDB", "FerretDB", opts)
		if err != nil {
			return nil, err
		}

		for _, issue := range issues {
			if issue.IsPullRequest() {
				prs = append(prs, issue)
			}
		}

		if resp.NextPage == 0 {
			return prs, nil
		}

		opts.ListOptions.Page = resp.NextPage
	}
}

// makeData creates template data.
func makeData(milestone *github.Milestone, prev string, prs []*github.Issue, l *slog.Logger) (*Data, error) {
	if milestone.ClosedAt != nil {
		l.Warn("milestone is closed")
	}

	d := &Data{
		PrevVersion: prev,
		Version:     *milestone.Title,
		Milestone:   *milestone.Number,
	}

	if milestone.DueOn == nil {
		return nil, fmt.Errorf("milestone %q does not have a due date", *milestone.Title)
	}

	d.Date = milestone.DueOn.Format("2006-01-02")

	var errs []error
	for _, pr := range prs {
		if pr.ClosedAt == nil {
			l.Warn(fmt.Sprintf("PR is not closed: %s", *pr.HTMLURL))
		}

		prData := PRData{
			Title:  *pr.Title,
			Author: *pr.User.Login,
			URL:    *pr.HTMLURL,
		}

		var found bool

		for _, label := range pr.Labels {
			switch *label.Name {
			case "code/feature":
				d.NewFeatures = append(d.NewFeatures, prData)
			case "code/bug", "code/bug-regression":
				d.FixedBugs = append(d.FixedBugs, prData)
			case "code/enhancement":
				d.Enhancements = append(d.Enhancements, prData)
			case "blog/engineering", "blog/marketing", "documentation":
				d.Documentation = append(d.Documentation, prData)
			case "code/chore", "project", "deps":
				d.OtherChanges = append(d.OtherChanges, prData)
			default:
				continue
			}

			if !found {
				found = true
				continue
			}

			errs = append(errs, fmt.Errorf("multiple possible categories for %s", prData.URL))
			break
		}

		if !found {
			errs = append(errs, fmt.Errorf("no category found for %s", prData.URL))
		}
	}

	return d, errors.Join(errs...)
}

// run generates the changelog.
func run(w io.Writer, l *slog.Logger, prev, next string) error {
	ctx := context.Background()

	client, err := gh.NewRESTClient(os.Getenv("GITHUB_TOKEN"), gh.SLogPrintf(l))
	if err != nil {
		return err
	}

	_, err = getMilestone(ctx, client, prev)
	if err != nil {
		return err
	}

	milestone, err := getMilestone(ctx, client, next)
	if err != nil {
		return err
	}

	l.Info("Received milestones", slog.Int("next", *milestone.Number))

	prs, err := getPRs(ctx, client, milestone)
	if err != nil {
		return err
	}

	l.Info("Received PRs", slog.Int("count", len(prs)))

	d, err := makeData(milestone, prev, prs, l)
	if err != nil {
		return err
	}

	t, err := template.New("").Option("missingkey=error").Parse(templateFile)
	if err != nil {
		return err
	}

	if err = t.Execute(w, d); err != nil {
		return err
	}

	return nil
}

func main() {
	prevF := flag.String("prev", "", "Previous version")
	nextF := flag.String("next", "", "Next version")
	flag.Parse()

	if *prevF == "" || *nextF == "" {
		log.Fatal("Both -prev and -next must be specified.")
	}

	var buf bytes.Buffer
	if err := run(&buf, slog.Default(), *prevF, *nextF); err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(os.Stdout, &buf); err != nil {
		log.Fatal(err)
	}
}
