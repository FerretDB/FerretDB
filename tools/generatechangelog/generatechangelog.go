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
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/FerretDB/gh"
	"github.com/google/go-github/v70/github"
)

//go:embed template.md
var template []byte

var categories = []struct {
	Name   string
	Labels map[string]struct{}
}{
	{
		Name: "NewFeatures",
		Labels: map[string]struct{}{
			"code/feature": {},
		},
	},
	{
		Name: "FixedBugs",
		Labels: map[string]struct{}{
			"code/bug":            {},
			"code/bug-regression": {},
		},
	},
	{
		Name: "Enhancements",
		Labels: map[string]struct{}{
			"code/enhancement": {},
		},
	},
	{
		Name: "Documentation",
		Labels: map[string]struct{}{
			"blog/engineering": {},
			"blog/marketing":   {},
			"documentation":    {},
		},
	},
	{
		Name: "OtherChanges",
		Labels: map[string]struct{}{
			"code/chore": {},
			"project":    {},
			"deps":       {},
		},
	},
}

type TemplateData struct {
	Categories map[string][]struct {
		Title  string
		Author string
		URL    string
	}
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
	}
}

func makeTemplateData(milestone *github.Milestone, prs []*github.Issue, l *slog.Logger) (*TemplateData, error) {
	return &TemplateData{}, nil
}

func run(w io.Writer, repoRoot, milestoneTitle, prev string) {
	ctx := context.Background()

	github.NewClient(p, log.Printf, gh.NoopPrintf, gh.NoopPrintf)

	client, err := gh.NewRESTClient(os.Getenv("GITHUB_TOKEN"), nil)
	if err != nil {
		log.Fatalf("Failed to create GitHub client: %v", err)
	}

	milestone, err := getMilestone(ctx, client, milestoneTitle)
	if err != nil {
		log.Fatalf("Failed to fetch milestone: %v", err)
	}

	mergedPRs, err := listMergedPRsOnMilestone(ctx, client, milestone)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}

	prs := groupPRsByCategories(mergedPRs, tpl.Changelog.Categories)

	tmpl, err := template.New("changelog").Parse(`
## [{{ .Current }}](https://github.com/FerretDB/FerretDB/releases/tag/{{ .Current }}) ({{ .Date }})
{{- $root := . }}
{{- range .Categories }}
{{ $prs := index $root.PRs . }}
{{- if $prs }}
### {{ . }}
{{ range $prs }}
- {{ .Title }} by @{{ .User }} in {{ .URL }}
{{- end }}
{{- end }}
{{- end }}
[All closed issues and pull requests]({{ .URL }}?closed=1).
{{- if .Previous }}
[All commits](https://github.com/FerretDB/FerretDB/compare/{{ .Previous }}...{{ .Current }}).
{{- end }}
`,
	)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	categories := make([]string, len(tpl.Changelog.Categories))
	for i, category := range tpl.Changelog.Categories {
		categories[i] = category.Title
	}

	data := struct { //nolint:vet // for readability
		Date       string
		Current    string
		Previous   string
		URL        string
		Categories []string
		PRs        map[string][]PR
	}{
		Date:       time.Now().Format("2006-01-02"),
		Current:    *milestone.Title,
		Previous:   prev,
		URL:        *milestone.HTMLURL,
		Categories: categories,
		PRs:        prs,
	}

	if err = tmpl.Execute(w, data); err != nil {
		log.Fatalf("Failed to render markdown: %v", err)
	}
}

func main() {
	prevF := flag.String("prev", "", "Previous milestone to compare against, e.g. v2.0.0-rc.1")
	nextF := flag.String("next", "Next", "Milestone to generate changelog for, e.g. v2.0.0-rc.2")
	flag.Parse()

	if *prevF == "" || *nextF == "" {
		log.Fatal("Both -prev and -next must be specified.")
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	run(os.Stdout, wd, *nextF, *prevF)
}
