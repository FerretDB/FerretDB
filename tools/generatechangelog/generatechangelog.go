package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/google/go-github/v57/github"
	"gopkg.in/yaml.v3"
)

const (
	OrgOwner = "FerretDB"
	Repo     = "FerretDB"
)

// The PR itself from the Github endpoint
type PRItem struct {
	URL    string `json:"html_url"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
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

func GetLatestTag(ctx context.Context, client *github.Client, owner, repo string) (*github.RepositoryTag, error) {
	tags, _, err := client.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{PerPage: 1})
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) == 0 {
		return nil, fmt.Errorf("no tags found in the repository")
	}

	return tags[0], nil
}

func GetCommitDate(ctx context.Context, client *github.Client, owner, repo, commitSHA string) (*time.Time, error) {
	commit, _, err := client.Git.GetCommit(ctx, owner, repo, commitSHA)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	return &commit.Author.Date.Time, nil
}

func FetchPRs(ctx context.Context, client *github.Client, owner, repo, sinceDate string) ([]PRItem, error) {
	query := fmt.Sprintf("repo:%s/%s is:pr is:merged merged:>=%s", owner, repo, sinceDate)
	opts := &github.SearchOptions{Sort: "created", Order: "desc", ListOptions: github.ListOptions{PerPage: 100}}

	issues, _, err := client.Search.Issues(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	res := []PRItem{}

	for _, issue := range issues.Issues {
		res = append(res, func(i *github.Issue) PRItem {
			labels := []struct {
				Name string "json:\"name\""
			}{}

			for _, label := range i.Labels {
				labels = append(labels, struct {
					Name string "json:\"name\""
				}{Name: *label.Name})
			}
			return PRItem{
				URL:    *i.URL,
				Number: *i.Number,
				Title:  *i.Title,
				User: struct {
					Login string "json:\"login\""
				}{
					Login: *i.User.Login,
				},
				Labels: labels,
			}
		}(issue))
	}

	return res, nil
}

// Loads the given release template
func LoadReleaseTemplate(filePath string) (*ReleaseTemplate, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var template *ReleaseTemplate
	err = yaml.Unmarshal(bytes, &template)
	if err != nil {
		return nil, err
	}

	return template, nil
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
	ctx := context.Background()

	client := NewGitHubClient()

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

	// Get latest tag
	latestTag, err := GetLatestTag(ctx, client, OrgOwner, Repo)
	if err != nil {
		log.Fatalf("Failed to get latest tag: %v", err)
	}

	commitDate, err := GetCommitDate(ctx, client, OrgOwner, Repo, *latestTag.Commit.SHA)
	if err != nil {
		log.Fatalf("Failed to get commit date of tag: %v", err)
	}

	sinceDate := commitDate.Format(time.RFC3339)

	// Fetch merged PRs for next version
	mergedPRs, err := FetchPRs(ctx, client, OrgOwner, Repo, sinceDate)
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
