// Package main contains script that generates changes for the latest version.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/google/go-github/v57/github"
	"gopkg.in/yaml.v3"
)

const (
	OrgOwner = "FerretDB"
	Repo     = "FerretDB"
)

// The PR itself from the Github endpoint
type PRItem struct {
	URL    string
	Number int
	Title  string
	User   struct {
		Login string
	}
	Labels []struct {
		Name string
	}
}

// The deconstructed template categories from the given template file
type ReleaseTemplate struct {
	Changelog struct {
		Categories []TemplateCategory `yaml:"categories"`
	} `yaml:"changelog"`
}

type TemplateCategory struct {
	Title  string   `yaml:"title"`
	Labels []string `yaml:"labels"`
}

// An intermediate struct to group PRs by labels and categories
type GroupedPRs struct {
	CategoryTitle string
	PRs           []PRItem
}

type CategorizedPRs struct {
	Groups []GroupedPRs
}

func NewGitHubClient() *github.Client {
	return github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_TOKEN"))
}

func GetMilestone(ctx context.Context, client *github.Client, milestoneTitle string) (*github.Milestone, error) {
	milestones, _, err := client.Issues.ListMilestones(ctx, "FerretDB", "FerretDB", &github.MilestoneListOptions{
		State: "all",
	})

	if err != nil {
		return nil, err
	}

	for _, milestone := range milestones {
		if *milestone.Title == milestoneTitle {
			return milestone, nil
		}
	}

	return nil, fmt.Errorf("no milestone found with the name %s", milestoneTitle)
}

func ListMergedPRsOnMilestone(ctx context.Context, client *github.Client, milestoneNumber int) ([]PRItem, error) {
	issues, _, err := client.Issues.ListByRepo(
		ctx,
		"FerretDB",
		"FerretDB",
		&github.IssueListByRepoOptions{
			State:     "closed",
			Milestone: strconv.Itoa(milestoneNumber),
		})

	if err != nil {
		return nil, err
	}

	var prItems []PRItem
	for _, issue := range issues {
		if !issue.IsPullRequest() {
			continue
		}

		labels := []struct {
			Name string
		}{}

		for _, label := range issue.Labels {
			labels = append(labels, struct{ Name string }{Name: *label.Name})
		}

		prItem := PRItem{
			URL:    *issue.URL,
			Number: *issue.Number,
			Title:  *issue.Title,
			User: struct{ Login string }{
				Login: *issue.User.Login,
			},
			Labels: labels,
		}

		prItems = append(prItems, prItem)
	}

	return prItems, nil
}

// Loads the given release template
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

// Helper function that generates slice of PRItems with input slice of PRs and set of labels
func collectPRItemsWithLabels(prItems []PRItem, labelSet map[string]struct{}) []PRItem {
	res := []PRItem{}

	for _, prItem := range prItems {
		for _, label := range prItem.Labels {
			if _, exists := labelSet[label.Name]; exists {
				res = append(res, prItem)
				break
			}
		}
	}

	return res
}

// Generating a Group of PRs based on the template category
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

// Iterating through the categories and generating Groups of PRs
func GroupPRsByCategories(prItems []PRItem, categories []TemplateCategory) CategorizedPRs {
	var categorizedPRs CategorizedPRs

	for _, category := range categories {
		grouped := groupPRsByTemplateCategory(prItems, category)
		if len(grouped.PRs) > 0 {
			categorizedPRs.Groups = append(categorizedPRs.Groups, *grouped)
		}
	}

	return categorizedPRs
}

// Rendering markdown based on template to stdout
func RenderMarkdownFromFile(categorizedPRs CategorizedPRs, templatePath string) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(os.Stdout, categorizedPRs); err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout)

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

	// Use existing release.yml file in .github
	releaseYamlFile := filepath.Join(repoRoot, ".github", "release.yml")

	// Load the release template file
	template, err := LoadReleaseTemplate(releaseYamlFile)
	if err != nil {
		log.Fatalf("Failed to read from template yaml file: %v", err)
	}

	ctx := context.Background()

	client := NewGitHubClient()

	milestone, err := GetMilestone(ctx, client, milestoneTitle)
	if err != nil {
		log.Fatalf("Failed to fetch milestone: %v", err)
	}

	// Fetch merged PRs for retrieved milestone
	mergedPRs, err := ListMergedPRsOnMilestone(ctx, client, *milestone.Number)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}

	// Group PRs by labels
	categorizedPRs := GroupPRsByCategories(mergedPRs, template.Changelog.Categories)

	// Render to markdown
	markdownTemplatePath := filepath.Join(repoRoot, "tools", "generatechangelog", "changelog_template.tmpl")
	if err := RenderMarkdownFromFile(categorizedPRs, markdownTemplatePath); err != nil {
		log.Fatalf("Failed to render markdown: %v", err)
	}
}
