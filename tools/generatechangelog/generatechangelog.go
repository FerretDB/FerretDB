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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/FerretDB/gh"
	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v57/github"
	"gopkg.in/yaml.v3"
)

// ReleaseTemplate represents GitHub release template.
type ReleaseTemplate struct {
	Changelog struct {
		Categories []TemplateCategory `yaml:"categories"`
	} `yaml:"changelog"`
}

// TemplateCategory represents a category in the template file.
type TemplateCategory struct {
	Title  string   `yaml:"title"`
	Labels []string `yaml:"labels"`
}

// GroupedPRs represented PRs grouped by categories.
type GroupedPRs struct {
	CategoryTitle string
	PRs           []PRItem
}

// PRItem represents GitHub PR.
type PRItem struct { //nolint:vet // for readability
	URL    string
	Title  string
	Number int
	User   string
	Labels []string
}

// getMilestone fetches the milestone with the given title.
func getMilestone(ctx context.Context, client *github.Client, milestoneTitle string) (current, previous *github.Milestone, err error) { //nolint:lll // for readability
	milestones, _, err := client.Issues.ListMilestones(ctx, "FerretDB", "FerretDB", &github.MilestoneListOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: 500,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	// Sort milestones by version number so we can find the previous milestone
	slices.SortFunc(milestones, compareMilestones)

	for _, milestone := range milestones {
		if *milestone.Title == milestoneTitle {
			current = milestone
			return
		}

		// todo: user semver lib to find previous ???
		previous = milestone
	}

	return nil, nil, fmt.Errorf("no milestone found with the name %s", milestoneTitle)
}

// compareMilestones compares two milestones by their version numbers.
// It returns a negative number when a < b, a positive number when
// a > b and zero when a == b or a and b are incomparable.
func compareMilestones(a, b *github.Milestone) int {
	aTitle := strings.Fields(*a.Title)[0]
	bTitle := strings.Fields(*b.Title)[0]

	aVer := semver.MustParse(aTitle)
	bVer := semver.MustParse(bTitle)

	return aVer.Compare(bVer)
}

// listMergedPRsOnMilestone returns the list of merged PRs on the given milestone.
func listMergedPRsOnMilestone(ctx context.Context, client *github.Client, milestone *github.Milestone) ([]PRItem, error) {
	issues, _, err := client.Issues.ListByRepo(
		ctx,
		"FerretDB",
		"FerretDB",
		&github.IssueListByRepoOptions{
			State:     "closed",
			Milestone: strconv.Itoa(*milestone.Number),
			Sort:      "created",
			Direction: "asc",
		})
	if err != nil {
		return nil, err
	}

	var prItems []PRItem

	for _, issue := range issues {
		if !issue.IsPullRequest() {
			continue
		}

		var labels []string

		for _, label := range issue.Labels {
			labels = append(labels, *label.Name)
		}

		prItem := PRItem{
			URL:    *issue.PullRequestLinks.HTMLURL,
			Number: *issue.Number,
			Title:  *issue.Title,
			User:   *issue.User.Login,
			Labels: labels,
		}

		prItems = append(prItems, prItem)
	}

	return prItems, nil
}

// loadReleaseTemplate loads the given release template.
func loadReleaseTemplate(filePath string) (*ReleaseTemplate, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tpl ReleaseTemplate

	err = yaml.Unmarshal(bytes, &tpl)
	if err != nil {
		return nil, err
	}

	return &tpl, nil
}

// collectPRItemsWithLabels generates slice of PRItems with input slice of PRs and set of labels.
func collectPRItemsWithLabels(prItems []PRItem, labelSet map[string]struct{}) []PRItem {
	var res []PRItem

	for _, prItem := range prItems {
		for _, label := range prItem.Labels {
			if _, exists := labelSet[label]; exists {
				res = append(res, prItem)
				break
			}
		}
	}

	return res
}

// groupPRsByTemplateCategory generates a group of PRs based on the template category.
func groupPRsByTemplateCategory(prItems []PRItem, templateCategory TemplateCategory) *GroupedPRs {
	labelSet := make(map[string]struct{})

	for _, label := range templateCategory.Labels {
		labelSet[label] = struct{}{}
	}

	return &GroupedPRs{
		CategoryTitle: templateCategory.Title,
		PRs:           collectPRItemsWithLabels(prItems, labelSet),
	}
}

// groupPRsByCategories iterates through the categories and generates Groups of PRs.
func groupPRsByCategories(prItems []PRItem, categories []TemplateCategory) []GroupedPRs {
	var categorizedPRs []GroupedPRs

	for _, category := range categories {
		grouped := groupPRsByTemplateCategory(prItems, category)
		if len(grouped.PRs) > 0 {
			categorizedPRs = append(categorizedPRs, *grouped)
		}
	}

	return categorizedPRs
}

func run(repoRoot, milestoneTitle string) {
	releaseYamlFile := filepath.Join(repoRoot, ".github", "release.yml")

	tpl, err := loadReleaseTemplate(releaseYamlFile)
	if err != nil {
		log.Fatalf("Failed to read from template yaml file: %v", err)
	}

	ctx := context.Background()

	client, err := gh.NewRESTClient(os.Getenv("GITHUB_TOKEN"), nil)
	if err != nil {
		log.Fatalf("Failed to create GitHub client: %v", err)
	}

	milestone, previous, err := getMilestone(ctx, client, milestoneTitle)
	if err != nil {
		log.Fatalf("Failed to fetch milestone: %v", err)
	}

	mergedPRs, err := listMergedPRsOnMilestone(ctx, client, milestone)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}

	categorizedPRs := groupPRsByCategories(mergedPRs, tpl.Changelog.Categories)

	templatePath := filepath.Join(repoRoot, "tools", "generatechangelog", "changelog_template.tmpl")

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Fatalf("Failed to parse template file: %v", err)
	}

	var previousTitle string

	if previous != nil {
		previousTitle = *previous.Title
		re := regexp.MustCompile(`^v\d+\.\d+\.\d+`)
		previousTitle = re.FindString(previousTitle)
	}

	currentTitle := *milestone.Title
	re := regexp.MustCompile(`^v\d+\.\d+\.\d+`)
	currentTitle = re.FindString(currentTitle)

	data := struct {
		Header   string
		Date     string
		Previous string
		Current  string
		URL      string
		PRs      []GroupedPRs
	}{
		Header:   *milestone.Title, // e.g. v0.8.0 Beta
		Date:     time.Now().Format("2006-01-02"),
		Previous: previousTitle, // e.g. v0.7.0
		Current:  currentTitle,  // e.g. v0.8.0
		URL:      *milestone.HTMLURL,
		PRs:      categorizedPRs,
	}

	if err = tmpl.Execute(os.Stdout, data); err != nil {
		log.Fatalf("Failed to render markdown: %v", err)
	}

	_, _ = fmt.Fprintln(os.Stdout)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: generatechangelog MILESTONE_TITLE")
		os.Exit(1)
	}
	milestoneTitle := os.Args[1]

	repoRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	run(repoRoot, milestoneTitle)
}
