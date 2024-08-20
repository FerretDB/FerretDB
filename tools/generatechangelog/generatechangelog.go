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
	"text/template"
	"time"

	"github.com/FerretDB/gh"
	"github.com/google/go-github/v57/github"
	"gopkg.in/yaml.v3"
)

// PRItem represents GitHub's PR.
type PRItem struct { //nolint:vet // for readability
	URL    string
	Title  string
	Number int
	User   string
	Labels []string
}

// ReleaseTemplate represents template categories from the given template file.
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

// GroupedPRs is an intermediate struct to group PRs by labels and categories.
type GroupedPRs struct {
	CategoryTitle string
	PRs           []PRItem
}

// contributors represents existing contributors.
var contributors = map[string]struct{}{}

// GetMilestone fetches the milestone with the given title.
func GetMilestone(ctx context.Context, client *github.Client, milestoneTitle string) (current, previous *github.Milestone, err error) {
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

		previous = milestone
	}

	return nil, nil, fmt.Errorf("no milestone found with the name %s", milestoneTitle)
}

// compareMilestones compares two milestones by their version numbers.
// It returns a negative number when a < b, a positive number when
// a > b and zero when a == b or a and b are incomparable
func compareMilestones(a, b *github.Milestone) int {
	re := regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)
	aMatches := re.FindStringSubmatch(*a.Title)
	bMatches := re.FindStringSubmatch(*b.Title)

	if len(aMatches) != 4 || len(bMatches) != 4 {
		return 0
	}

	aMajor, _ := strconv.Atoi(aMatches[1])
	aMinor, _ := strconv.Atoi(aMatches[2])
	aPatch, _ := strconv.Atoi(aMatches[3])

	bMajor, _ := strconv.Atoi(bMatches[1])
	bMinor, _ := strconv.Atoi(bMatches[2])
	bPatch, _ := strconv.Atoi(bMatches[3])

	if aMajor != bMajor {
		return aMajor - bMajor
	}

	if aMinor != bMinor {
		return aMinor - bMinor
	}

	if aPatch != bPatch {
		return aPatch - bPatch
	}

	return 0
}

// ListMergedPRsOnMilestone returns the list of merged PRs on the given milestone and the set of first time contributors with their first contributions.
func ListMergedPRsOnMilestone(ctx context.Context, client *github.Client, milestone *github.Milestone) ([]PRItem, map[string]string, error) {
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
		return nil, nil, err
	}

	var prItems []PRItem
	contributors := make(map[string]string)

	for _, issue := range issues {
		if !issue.IsPullRequest() {
			continue
		}

		var labels []string

		for _, label := range issue.Labels {
			labels = append(labels, *label.Name)
		}

		prItem := PRItem{
			URL:    *issue.URL,
			Number: *issue.Number,
			Title:  *issue.Title,
			User:   *issue.User.Login,
			Labels: labels,
		}

		if _, exists := contributors[prItem.User]; !exists {
			contributors[prItem.User] = prItem.URL
		}

		prItems = append(prItems, prItem)
	}

	for user := range contributors {
		uIssues, _, err := client.Issues.ListByRepo(
			ctx,
			"FerretDB",
			"FerretDB",
			&github.IssueListByRepoOptions{
				State:   "closed",
				Creator: user,
			})
		if err != nil {
			return nil, nil, err
		}

		fmt.Printf("current milestone: %s\n", *milestone.Title)

		for _, issue := range uIssues {
			if !issue.IsPullRequest() || issue.Milestone == nil || compareMilestones(issue.Milestone, milestone) <= 0 {
				continue
			}

			fmt.Printf("contrubutor: %s, issue: %d, milestone: %s\n", user, *issue.Number, *issue.Milestone.Title)

			delete(contributors, user)
			break
		}
	}

	fmt.Printf("new contribs: %d", len(contributors))

	return prItems, contributors, nil
}

// LoadReleaseTemplate loads the given release template.
func LoadReleaseTemplate(filePath string) (*ReleaseTemplate, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var template ReleaseTemplate

	err = yaml.Unmarshal(bytes, &template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

// collectPRItemsWithLabels generates slice of PRItems with input slice of PRs and set of labels.
func collectPRItemsWithLabels(prItems []PRItem, labelSet map[string]struct{}) []PRItem {
	res := []PRItem{}

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

	// Generate set of labels to check against PR
	for _, label := range templateCategory.Labels {
		labelSet[label] = struct{}{}
	}

	return &GroupedPRs{
		CategoryTitle: templateCategory.Title,
		PRs:           collectPRItemsWithLabels(prItems, labelSet),
	}
}

// GroupPRsByCategories iterates through the categories and generates Groups of PRs.
func GroupPRsByCategories(prItems []PRItem, categories []TemplateCategory) []GroupedPRs {
	var categorizedPRs []GroupedPRs

	for _, category := range categories {
		grouped := groupPRsByTemplateCategory(prItems, category)
		if len(grouped.PRs) > 0 {
			categorizedPRs = append(categorizedPRs, *grouped)
		}
	}

	return categorizedPRs
}

// RenderMarkdownFromFile renders markdown based on template to stdout.
func RenderMarkdownFromFile(categorizedPRs []GroupedPRs, templatePath string) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	if err = tmpl.Execute(os.Stdout, categorizedPRs); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(os.Stdout)

	return nil
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

	releaseYamlFile := filepath.Join(repoRoot, ".github", "release.yml")

	// Load the release template file
	tpl, err := LoadReleaseTemplate(releaseYamlFile)
	if err != nil {
		log.Fatalf("Failed to read from template yaml file: %v", err)
	}

	ctx := context.Background()

	client, err := gh.NewRESTClient(os.Getenv("GITHUB_TOKEN"), nil)
	if err != nil {
		log.Fatalf("Failed to create GitHub client: %v", err)
	}

	milestone, previous, err := GetMilestone(ctx, client, milestoneTitle)
	if err != nil {
		log.Fatalf("Failed to fetch milestone: %v", err)
	}

	// Fetch merged PRs for retrieved milestone
	mergedPRs, firstTimers, err := ListMergedPRsOnMilestone(ctx, client, milestone)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}

	// Group PRs by labels
	categorizedPRs := GroupPRsByCategories(mergedPRs, tpl.Changelog.Categories)

	// Render to markdown
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
		Header      string
		Date        string
		Previous    string
		Current     string
		URL         string
		PRs         []GroupedPRs
		FirstTimers map[string]string
	}{
		Header:      *milestone.Title, // e.g. v0.8.0 Beta
		Date:        time.Now().Format("2006-01-02"),
		Previous:    previousTitle, // e.g. v0.7.0
		Current:     currentTitle,  // e.g. v0.8.0
		URL:         *milestone.HTMLURL,
		PRs:         categorizedPRs,
		FirstTimers: firstTimers,
	}

	if err = tmpl.Execute(os.Stdout, data); err != nil {
		log.Fatalf("Failed to render markdown: %v", err)
	}

	_, _ = fmt.Fprintln(os.Stdout)
}
