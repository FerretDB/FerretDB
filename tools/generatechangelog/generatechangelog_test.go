package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TagDate = "2024-03-20"
)

func TestGetLatestTagAndCommitDate(t *testing.T) {
	ctx := context.Background()
	client := NewGitHubClient()

	latestTag, err := GetLatestTag(ctx, client, OrgOwner, Repo)
	require.NoError(t, err)

	tagName := ""
	if latestTag.Name != nil {
		tagName = *latestTag.Name
	}

	commitDate, err := GetCommitDate(ctx, client, OrgOwner, Repo, *latestTag.Commit.SHA)
	require.NoError(t, err)

	t.Logf("Latest Tag: %s, Commit Date: %s", tagName, commitDate.Format(time.RFC3339))
}

func TestFetchPRs(t *testing.T) {
	ctx := context.Background()
	client := NewGitHubClient()

	mergedPRs, err := FetchPRs(ctx, client, OrgOwner, Repo, TagDate)
	require.NoError(t, err)

	responseJSON, err := json.MarshalIndent(mergedPRs, "", "  ")
	require.NoError(t, err)
	t.Logf("Fetched PRs: %s", responseJSON)
}

func TestLoadReleaseTemplate(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	releaseYamlFile := filepath.Join(cwd, "test_template.yml")

	template, err := LoadReleaseTemplate(releaseYamlFile)

	require.NoError(t, err)

	expectedNumCategories := 3
	assert.Len(t, template.Changelog.Categories, expectedNumCategories, fmt.Sprintf("Expected %d categories", expectedNumCategories))

	expectedCategories := []TemplateCategory{
		{Title: "features", Labels: []string{"code/feature"}},
		{Title: "bugs", Labels: []string{"code/bug"}},
		{Title: "enhancements", Labels: []string{"code/enhancement"}},
	}

	for i, category := range template.Changelog.Categories {
		assert.Equal(t, expectedCategories[i].Title, category.Title, "Expected category title to match")
		assert.Equal(t, len(expectedCategories[i].Labels), len(category.Labels), "Expected label length to match")
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

	assert.Equal(t, expected, actual, "Expected groups to be equal")
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
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(templateContent))
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

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
	require.NoError(t, err)

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	expectedOutput := "\n### Features 🎉\n- Add feature X by @dev1 in https://github.com/FerretDB/FerretDB/pull/1\n### Bugs 🐛\n- Fix bug Y by @dev2 in https://github.com/FerretDB/FerretDB/pull/2\n"

	assert.Equal(t, expectedOutput, buf.String(), "Expected rendered markdown to be equal")
}
