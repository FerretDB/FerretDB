package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v3"
)

const (
	GITHUB_HEADER = "application/vnd.github+json"
	ORG_NAME      = "FerretDB"
	REPO_NAME     = "FerretDB"
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

// The response from the endpoint
type PRResponse struct {
	TotalCount        int      `json:"total_count"`
	IncompleteResults bool     `json:"incomplete_results"`
	Items             []PRItem `json:"items"`
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

// Opens the current repository
func GetRepository(repoPath string) (*git.Repository, error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return r, nil
}

// Gets the latest tag available locally. This should correspond with the
// version.
func GetLatestTag(repo *git.Repository) (string, time.Time, error) {
	tagRefs, err := repo.Tags()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get tags: %w", err)
	}

	var latestTagCommitTime time.Time
	var latestTagName string

	err = tagRefs.ForEach(func(t *plumbing.Reference) error {
		obj, err := repo.TagObject(t.Hash())
		var commitHash plumbing.Hash
		if err == nil {
			commitHash = obj.Target
		} else {
			commitHash = t.Hash()
		}

		commit, err := repo.CommitObject(commitHash)
		if err != nil {
			return fmt.Errorf("failed to get commit object: %w", err)
		}

		if commit.Committer.When.After(latestTagCommitTime) {
			latestTagCommitTime = commit.Committer.When
			latestTagName = t.Name().Short()
		}

		return nil
	})

	if err != nil {
		return "", time.Time{}, fmt.Errorf("error iterating tags: %w", err)
	}

	if latestTagName == "" {
		return "", time.Time{}, fmt.Errorf("no tags found in the repository")
	}

	return latestTagName, latestTagCommitTime, nil
}

// Get the endpoint, with param to obtain PRs merged past input date
func GetGithubPRUrl(orgName, repoName, date string) string {
	return fmt.Sprintf("https://api.github.com/search/issues?q=repo:%s/%s+is:pr+is:merged+merged:>=%s", orgName, repoName, date)
}

// Fetches the PRs and unmarshals into a PRResponse
func FetchPRs(client *http.Client, prUrl string) (*PRResponse, error) {
	req, err := http.NewRequest("GET", prUrl, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", GITHUB_HEADER)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var prResponse PRResponse

	if err := json.Unmarshal(body, &prResponse); err != nil {
		return nil, err
	}

	return &prResponse, nil
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

	for _, label := range templateCategory.Labels {
		labelSet[label] = struct{}{}
	}

	return &GroupedPRs{
		CategoryTitle: templateCategory.Title,
		PRs:           collectPRItemsWithLabels(prItems, labelSet),
	}
}

// Iterating throught the categories and generating Groups of PRs
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
	repoRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	repo, err := GetRepository(repoRoot)

	if err != nil {
		log.Fatalf("Failed to open repo: %v", err)
	}

	releaseYamlFile := filepath.Join(repoRoot, ".github", "release.yml")

	template, err := LoadReleaseTemplate(releaseYamlFile)

	if err != nil {
		log.Fatalf("Failed to read from template yaml file: %v", err)
	}

	_, latestTagDate, err := GetLatestTag(repo)
	if err != nil {
		log.Fatalf("Failed to get latest tag: %v", err)
	}

	dateString := latestTagDate.Format("2006-01-02")

	prUrl := GetGithubPRUrl(ORG_NAME, REPO_NAME, dateString)

	client := &http.Client{}
	prResponse, err := FetchPRs(client, prUrl)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}

	categorizedPRs := GroupPRsByCategories(prResponse.Items, template.Changelog.Categories)

	markdownTemplatePath := filepath.Join(repoRoot, "tools", "generatechangelog", "changelog_template.tmpl")
	if err := RenderMarkdownFromFile(categorizedPRs, markdownTemplatePath); err != nil {
		log.Fatalf("Failed to render markdown: %v", err)
	}
}
