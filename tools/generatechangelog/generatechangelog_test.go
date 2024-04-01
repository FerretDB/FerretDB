package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

const (
	TAG_NAME = "v1.21.0"
	TAG_DATE = "2024-03-20"
)

func TestGetLatestTag(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	repoRoot := filepath.Join(cwd, "../..")

	repo, err := GetRepository(repoRoot)

	if err != nil {
		t.Fatalf("Could not open repo: %v", err)
	}

	tagName, tagDate, err := GetLatestTag(repo)
	if err != nil {
		t.Fatalf("GetLatestTag returned an unexpected error: %v", err)
	}

	if tagName == "" {
		t.Error("Expected a non-empty tag name")
	}

	if tagDate.IsZero() {
		t.Error("Expected a non-zero tag date")
	}

	t.Logf("Latest Tag: %s, Date: %s", tagName, tagDate.Format(time.RFC3339))
}

func TestFetchPRs(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PRResponse{
			TotalCount:        2,
			IncompleteResults: false,
			Items: []PRItem{
				{
					URL:    "http://example.com/pr1",
					Number: 1,
					Title:  "First PR",
					User: struct {
						Login string `json:"login"`
					}{
						Login: "firstuser",
					},
					Labels: []struct {
						Name string `json:"name"`
					}{
						{Name: "bug"},
						{Name: "good first issue"},
					},
				},
				{
					URL:    "http://example.com/pr2",
					Number: 2,
					Title:  "Second PR",
					User: struct {
						Login string `json:"login"`
					}{
						Login: "seconduser",
					},
					Labels: []struct {
						Name string `json:"name"`
					}{
						{Name: "enhancement"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	prURL := mockServer.URL

	client := &http.Client{}
	prResponse, err := FetchPRs(client, prURL)
	if err != nil {
		t.Fatalf("FetchPRs failed: %v", err)
	}

	if prResponse.TotalCount != 2 {
		t.Errorf("Expected TotalCount to be 2, got %d", prResponse.TotalCount)
	}

	if prResponse.Items[0].User.Login != "firstuser" {
		t.Errorf("Expected User.Login of the first PR to be 'firstuser', got '%s'", prResponse.Items[0].User.Login)
	}
	if len(prResponse.Items[0].Labels) != 2 {
		t.Errorf("Expected 2 labels for the first PR, got %d", len(prResponse.Items[0].Labels))
	}
}

func TestFetchPRsIntegration(t *testing.T) {
	prUrl := GetGithubPRUrl(ORG_NAME, REPO_NAME, TAG_DATE)

	client := &http.Client{}

	prResponse, err := FetchPRs(client, prUrl)

	if err != nil {
		t.Fatalf("FetchPRs failed: %v", err)
	}

	responseJSON, err := json.MarshalIndent(prResponse, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal PRResponse for logging: %v", err)
	}
	t.Logf("Fetched PRs: %s", responseJSON)
}

func TestLoadReleaseTemplate(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	releaseYamlFile := filepath.Join(cwd, "test_template.yml")

	template, err := LoadReleaseTemplate(releaseYamlFile)

	if err != nil {
		t.Fatalf("Failed to read from template yaml file: %v", err)
	}

	expectedNumCategories := 3
	if len(template.Changelog.Categories) != expectedNumCategories {
		t.Errorf("Expected %d categories, got %d", expectedNumCategories, len(template.Changelog.Categories))
	}

	expectedCategories := []TemplateCategory{
		{Title: "features", Labels: []string{"code/feature"}},
		{Title: "bugs", Labels: []string{"code/bug"}},
		{Title: "enhancements", Labels: []string{"code/enhancement"}},
	}

	for i, category := range template.Changelog.Categories {
		if category.Title != expectedCategories[i].Title {
			t.Errorf("Category title mismatch, expected: %s, got: %s", expectedCategories[i].Title, category.Title)
		}
		if len(category.Labels) != len(expectedCategories[i].Labels) {
			t.Fatalf("Labels length mismatch in category '%s', expected: %d, got: %d", category.Title, len(expectedCategories[i].Labels), len(category.Labels))
		}
		for j, label := range category.Labels {
			if label != expectedCategories[i].Labels[j] {
				t.Errorf("Label mismatch in category '%s', expected: %s, got: %s", category.Title, expectedCategories[i].Labels[j], label)
			}
		}
	}
}

func TestGroupPRsByCategories(t *testing.T) {
	prItems := []PRItem{
		{
			URL:    "http://example.com/pr1",
			Number: 1,
			Title:  "First PR",
			User: struct {
				Login string `json:"login"`
			}{Login: "user1"},
			Labels: []struct {
				Name string `json:"name"`
			}{{Name: "code/feature"}},
		},
		{
			URL:    "http://example.com/pr2",
			Number: 2,
			Title:  "Second PR",
			User: struct {
				Login string `json:"login"`
			}{Login: "user2"},
			Labels: []struct {
				Name string `json:"name"`
			}{{Name: "code/bug"}},
		},
		{
			URL:    "http://example.com/pr3",
			Number: 3,
			Title:  "Third PR",
			User: struct {
				Login string `json:"login"`
			}{Login: "user3"},
			Labels: []struct {
				Name string `json:"name"`
			}{{Name: "code/enhancement"}},
		},
	}

	categories := []TemplateCategory{
		{Title: "Features", Labels: []string{"code/feature"}},
		{Title: "Bugs", Labels: []string{"code/bug"}},
		{Title: "Enhancements", Labels: []string{"code/enhancement"}},
	}

	expected := CategorizedPRs{
		Groups: []GroupedPRs{
			{
				CategoryTitle: "Features",
				PRs:           []PRItem{prItems[0]},
			},
			{
				CategoryTitle: "Bugs",
				PRs:           []PRItem{prItems[1]},
			},
			{
				CategoryTitle: "Enhancements",
				PRs:           []PRItem{prItems[2]},
			},
		},
	}

	actual := GroupPRsByCategories(prItems, categories)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Grouping PRs by categories did not match expected result.\nExpected: %+v\nGot: %+v", expected, actual)
	}
}

func TestRenderMarkdownFromFile(t *testing.T) {
	// redirect os.Stdout to a buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	// create a temporary template file to remove later
	templateContent := `{{- range .Groups }}
### {{ .CategoryTitle }}

{{- range .PRs }}
- {{ .Title }} by @{{ .User.Login }} in {{ .URL }}
{{- end }}
{{- end }}`

	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(templateContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	categorizedPRs := CategorizedPRs{
		Groups: []GroupedPRs{
			{
				CategoryTitle: "Features 🎉",
				PRs: []PRItem{
					{
						URL:    "https://github.com/FerretDB/FerretDB/pull/1",
						Number: 1,
						Title:  "Add feature X",
						User: struct {
							Login string `json:"login"`
						}{Login: "dev1"},
						Labels: []struct {
							Name string `json:"name"`
						}{{Name: "code/feature"}},
					},
				},
			},
			{
				CategoryTitle: "Bugs 🐛",
				PRs: []PRItem{
					{
						URL:    "https://github.com/FerretDB/FerretDB/pull/2",
						Number: 2,
						Title:  "Fix bug Y",
						User: struct {
							Login string `json:"login"`
						}{Login: "dev2"},
						Labels: []struct {
							Name string `json:"name"`
						}{{Name: "code/bug"}},
					},
				},
			},
		},
	}
	err = RenderMarkdownFromFile(categorizedPRs, tmpfile.Name())
	if err != nil {
		t.Fatalf("RenderMarkdownFromFile returned an error: %v", err)
	}

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	expectedOutput := "\n### Features 🎉\n- Add feature X by @dev1 in https://github.com/FerretDB/FerretDB/pull/1\n### Bugs 🐛\n- Fix bug Y by @dev2 in https://github.com/FerretDB/FerretDB/pull/2\n"

	if buf.String() != expectedOutput {
		t.Errorf("Expected output to be %q, got %q", expectedOutput, buf.String())
	}
}
