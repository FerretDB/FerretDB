package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TagDate = "2024-03-20"
)

func TestGetMilestone(t *testing.T) {
	ctx := context.Background()
	client := NewGitHubClient()

	milestoneTitle := "v0.9.1"

	milestone, err := GetMilestone(ctx, client, milestoneTitle)
	require.NoError(t, err)
	require.NotNil(t, milestone, "The milestone should not be nil")
	require.Equal(t, milestoneTitle, *milestone.Title, "Milestone title does not match")
	require.Equal(t, 30, *milestone.Number, "Milestone Number does not match")
	require.Equal(t, "closed", *milestone.State, "Milestone should be closed")
	require.Equal(t, 29, *milestone.ClosedIssues, "The number of closed issues does not match")

	t.Logf("Milestone details:\n- Title: %s\n- Number: %d\n- State: %s\n- Closed Issues: %d\n- Description: %s",
		*milestone.Title,
		*milestone.Number,
		*milestone.State,
		*milestone.ClosedIssues,
		*milestone.Description)
}

func TestListMergedPRsOnMilestone(t *testing.T) {
	ctx := context.Background()
	client := NewGitHubClient()

	// The milestone number for "v0.9.1"
	milestoneNumber := 30

	prItems, err := ListMergedPRsOnMilestone(ctx, client, milestoneNumber)
	require.NoError(t, err)

	expectedNumberOfPRs := 21
	require.Len(t, prItems, expectedNumberOfPRs, "The number of PR items does not match the expected")

	if len(prItems) > 0 {
		t.Logf("PR items for milestone %d:\n", milestoneNumber)
		for _, prItem := range prItems {
			t.Logf("- PR #%d: %s by %s\n", prItem.Number, prItem.Title, prItem.User.Login)
			t.Logf("  URL: %s\n", prItem.URL)
			if len(prItem.Labels) > 0 {
				t.Log("  Labels:")
				for _, label := range prItem.Labels {
					t.Logf("    - %s\n", label.Name)
				}
			}
		}
	} else {
		t.Log("No PR items found for the milestone.")
	}
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
				Login string
			}{Login: "user1"},
			Labels: []struct {
				Name string
			}{{Name: "code/feature"}},
		},
		{
			URL:    "http://example.com/pr2",
			Number: 2,
			Title:  "Second PR",
			User: struct {
				Login string
			}{Login: "user2"},
			Labels: []struct {
				Name string
			}{{Name: "code/bug"}},
		},
		{
			URL:    "http://example.com/pr3",
			Number: 3,
			Title:  "Third PR",
			User: struct {
				Login string
			}{Login: "user3"},
			Labels: []struct {
				Name string
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
							Login string
						}{Login: "dev1"},
						Labels: []struct {
							Name string
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
							Login string
						}{Login: "dev2"},
						Labels: []struct {
							Name string
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

func TestGenerateChangelogIntegration(t *testing.T) {
	milestoneTitle := "v0.9.1"
	cwd, err := os.Getwd()
	require.NoError(t, err)
	releaseYamlFile := filepath.Join(cwd, "..", "..", ".github", "release.yml")
	template, err := LoadReleaseTemplate(releaseYamlFile)
	require.NoError(t, err)
	expectedNumCategories := 6
	assert.Len(t, template.Changelog.Categories, expectedNumCategories, fmt.Sprintf("Expected %d categories", expectedNumCategories))

	ctx := context.Background()
	client := NewGitHubClient()

	milestone, err := GetMilestone(ctx, client, milestoneTitle)
	require.NoError(t, err)
	require.NotNil(t, milestone, "The milestone should not be nil")
	require.Equal(t, milestoneTitle, *milestone.Title, "Milestone title does not match")
	require.Equal(t, 30, *milestone.Number, "Milestone Number does not match")
	require.Equal(t, "closed", *milestone.State, "Milestone should be closed")
	require.Equal(t, 29, *milestone.ClosedIssues, "The number of closed issues does not match")

	prItems, err := ListMergedPRsOnMilestone(ctx, client, *milestone.Number)
	expectedNumberOfPRs := 21
	require.Len(t, prItems, expectedNumberOfPRs, "The number of PR items does not match the expected")

	categorizedPRs := GroupPRsByCategories(prItems, template.Changelog.Categories)

	mdTemplate := filepath.Join(cwd, "changelog_template.tmpl")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	err = RenderMarkdownFromFile(categorizedPRs, mdTemplate)
	require.NoError(t, err)

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	expectedOutput := "\n### New Features 🎉\n\n- Support `listIndexes` command by @rumyantseva in https://api.github.com/repos/FerretDB/FerretDB/issues/1960\n- Pushdown Tigris queries with dot notation by @noisersup in https://api.github.com/repos/FerretDB/FerretDB/issues/1908\n- Support Tigris pushdowns for numbers by @noisersup in https://api.github.com/repos/FerretDB/FerretDB/issues/1842\n\n### Fixed Bugs 🐛\n\n- Fix key ordering on document replacing by @noisersup in https://api.github.com/repos/FerretDB/FerretDB/issues/1946\n- Fix SASL response for `PLAIN` authentication by @b1ron in https://api.github.com/repos/FerretDB/FerretDB/issues/1942\n- Fix `$pop` operator error handling of non-existent path by @chilagrow in https://api.github.com/repos/FerretDB/FerretDB/issues/1907\n\n### Documentation 📄\n\n- Prepare v0.9.1 release by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1958\n- Fix broken link by @Fashander in https://api.github.com/repos/FerretDB/FerretDB/issues/1918\n- Add blog post on \"MongoDB Alternatives: 5 Database Alternatives to MongoDB for 2023\" by @Fashander in https://api.github.com/repos/FerretDB/FerretDB/issues/1911\n- Bump deps by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1902\n\n### Other Changes 🤖\n\n- Prepare v0.9.1 release by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1958\n- Remove `skipTigrisPushdown` from tests by @noisersup in https://api.github.com/repos/FerretDB/FerretDB/issues/1957\n- Rename function, add TODO by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1955\n- Tweak CI settings by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1948\n- Add `iterator.WithClose` helper by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1947\n- Implement Tigris query iterator by @w84thesun in https://api.github.com/repos/FerretDB/FerretDB/issues/1924\n- Remove unused parameter by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1919\n- Bump Tigris by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1916\n- Assorted internal tweaks by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1909\n- Bump deps by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1902\n- Use multiple Tigris instances to run tests by @chilagrow in https://api.github.com/repos/FerretDB/FerretDB/issues/1878\n- Add simple `otel` tracing to collect data from tests by @rumyantseva in https://api.github.com/repos/FerretDB/FerretDB/issues/1863\n- Rework on integration test setup by @chilagrow in https://api.github.com/repos/FerretDB/FerretDB/issues/1857\n\n"

	assert.Equal(t, expectedOutput, buf.String(), "Expected rendered markdown to be equal")
}
