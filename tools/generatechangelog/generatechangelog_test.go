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

package main

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/FerretDB/gh"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/AlekSi/pointer"

	"github.com/google/go-github/v57/github"
)

func TestCompareMilestones(t *testing.T) {
	for name, tc := range map[string]struct {
		name     string
		a        *github.Milestone
		b        *github.Milestone
		expected int
	}{
		"v0.9.1 vs v0.9.0": {
			a:        &github.Milestone{Title: pointer.To("v0.9.1")},
			b:        &github.Milestone{Title: pointer.To("v0.9.0")},
			expected: 1,
		},
		"v0.9.0 vs v0.9.1": {
			a:        &github.Milestone{Title: pointer.To("v0.9.0")},
			b:        &github.Milestone{Title: pointer.To("v0.9.1")},
			expected: -1,
		},
		"v0.9.0 vs v0.9.0": {
			a:        &github.Milestone{Title: pointer.To("v0.9.0")},
			b:        &github.Milestone{Title: pointer.To("v0.9.0")},
			expected: 0,
		},
		"v0.9.1 vs v0.9.0 Developer Preview": {
			a:        &github.Milestone{Title: pointer.To("v0.9.1")},
			b:        &github.Milestone{Title: pointer.To("v0.9.0 Developer Preview")},
			expected: 1,
		},
		"v0.9.0 Developer Preview vs v0.9.1": {
			a:        &github.Milestone{Title: pointer.To("v0.9.0 Developer Preview")},
			b:        &github.Milestone{Title: pointer.To("v0.9.1")},
			expected: -1,
		},
		"v1.2.0 vs v1.21.0": {
			a:        &github.Milestone{Title: pointer.To("v1.2.0")},
			b:        &github.Milestone{Title: pointer.To("v1.21.0")},
			expected: -19,
		},
		"v0.2.0 vs v1.2.0": {
			a:        &github.Milestone{Title: pointer.To("v0.2.0")},
			b:        &github.Milestone{Title: pointer.To("v1.2.0")},
			expected: -1,
		},
		"v1.2.10 vs v1.2.1": {
			a:        &github.Milestone{Title: pointer.To("v1.2.10")},
			b:        &github.Milestone{Title: pointer.To("v1.2.1")},
			expected: 9,
		},
	} {
		t.Run(name, func(t *testing.T) {
			actual := compareMilestones(tc.a, tc.b)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetMilestone(t *testing.T) {
	ctx := context.Background()
	client, err := gh.NewRESTClient(os.Getenv("GITHUB_TOKEN"), log.Printf)
	require.NoError(t, err)

	milestoneTitle := "v0.9.1"

	milestone, previous, err := getMilestone(ctx, client, milestoneTitle)
	require.NoError(t, err)

	expectedMilestone := &github.Milestone{
		Title:        pointer.To("v0.9.1"),
		Number:       pointer.To(30),
		State:        pointer.To("closed"),
		ClosedIssues: pointer.To(29),
		Description:  pointer.To(""),
		URL:          pointer.To("https://api.github.com/repos/FerretDB/FerretDB/milestones/30"),
		HTMLURL:      pointer.To("https://github.com/FerretDB/FerretDB/milestone/30"),
		LabelsURL:    pointer.To("https://api.github.com/repos/FerretDB/FerretDB/milestones/30/labels"),
		ID:           pointer.To(int64(8941887)),
	}
	actualMilestone := &github.Milestone{
		Title:        milestone.Title,
		Number:       milestone.Number,
		State:        milestone.State,
		ClosedIssues: milestone.ClosedIssues,
		Description:  milestone.Description,
		URL:          milestone.URL,
		HTMLURL:      milestone.HTMLURL,
		LabelsURL:    milestone.LabelsURL,
		ID:           milestone.ID,
	}
	assert.Equal(t, expectedMilestone, actualMilestone)

	expectedPrevious := &github.Milestone{
		Title:        pointer.To("v0.9.0 Developer Preview"),
		Number:       pointer.To(13),
		State:        pointer.To("closed"),
		ClosedIssues: pointer.To(58),
		Description:  pointer.To(""),
		URL:          pointer.To("https://api.github.com/repos/FerretDB/FerretDB/milestones/13"),
		HTMLURL:      pointer.To("https://github.com/FerretDB/FerretDB/milestone/13"),
		LabelsURL:    pointer.To("https://api.github.com/repos/FerretDB/FerretDB/milestones/13/labels"),
		ID:           pointer.To(int64(7815597)),
	}
	actualPrevious := &github.Milestone{
		Title:        previous.Title,
		Number:       previous.Number,
		State:        previous.State,
		ClosedIssues: previous.ClosedIssues,
		Description:  previous.Description,
		URL:          previous.URL,
		HTMLURL:      previous.HTMLURL,
		LabelsURL:    previous.LabelsURL,
		ID:           previous.ID,
	}
	assert.Equal(t, expectedPrevious, actualPrevious)
}

/*
func TestListMergedPRsOnMilestone(t *testing.T) {
	ctx := context.Background()
	client, err := gh.NewRESTClient(os.Getenv("GITHUB_TOKEN"), log.Printf)
	require.NoError(t, err)

	// The milestone number for "v0.9.1"
	milestoneNumber := 30

	prItems, err := listMergedPRsOnMilestone(ctx, client, milestoneNumber)
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

	template, err := loadReleaseTemplate(releaseYamlFile)

	require.NoError(t, err)

	expectedNumCategories := 3
	assert.Len(t, template.Changelog.Categories, expectedNumCategories,
		fmt.Sprintf("Expected %d categories", expectedNumCategories))

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
			Labels: []string}{{Name: "code/feature"}},
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
	defer assert.NoError(t, os.Remove(tmpfile.Name()))

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

	assert.NoError(t, w.Close())
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	assert.NoError(t, err)

	expectedOutput := "\n### Features 🎉\n- Add feature X by @dev1 in https://github.com/FerretDB/FerretDB/pull/1\n" +
		"### Bugs 🐛\n- Fix bug Y by @dev2 in https://github.com/FerretDB/FerretDB/pull/2\n"

	assert.Equal(t, expectedOutput, buf.String(), "Expected rendered markdown to be equal")
}

/*
func TestGenerateChangelogIntegration(t *testing.T) {
	milestoneTitle := "v0.9.1"
	cwd, err := os.Getwd()
	require.NoError(t, err)
	releaseYamlFile := filepath.Join(cwd, "..", "..", ".github", "release.yml")
	template, err := LoadReleaseTemplate(releaseYamlFile)
	require.NoError(t, err)
	expectedNumCategories := 6
	assert.Len(t, template.Changelog.Categories, expectedNumCategories,
		fmt.Sprintf("Expected %d categories", expectedNumCategories))

	ctx := context.Background()
	client, err := gh.NewRESTClient(os.Getenv("GITHUB_TOKEN"), log.Printf)
	require.NoError(t, err)

	milestone, err := GetMilestone(ctx, client, milestoneTitle)
	require.NoError(t, err)
	require.NotNil(t, milestone, "The milestone should not be nil")
	require.Equal(t, milestoneTitle, *milestone.Title, "Milestone title does not match")
	require.Equal(t, 30, *milestone.Number, "Milestone Number does not match")
	require.Equal(t, "closed", *milestone.State, "Milestone should be closed")
	require.Equal(t, 29, *milestone.ClosedIssues, "The number of closed issues does not match")

	prItems, err := ListMergedPRsOnMilestone(ctx, client, *milestone.Number)
	require.NoError(t, err)
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

	assert.NoError(t, w.Close())
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	assert.NoError(t, err)

	expectedOutput := "\n### New Features 🎉\n\n" +
		"- Support `listIndexes` command by @rumyantseva in https://api.github.com/repos/FerretDB/FerretDB/issues/1960\n" +
		"- Pushdown Tigris queries with dot notation by @noisersup in " +
		"https://api.github.com/repos/FerretDB/FerretDB/issues/1908\n" +
		"- Support Tigris pushdowns for numbers by @noisersup in https://api.github.com/repos/FerretDB/FerretDB/issues/1842\n\n" +
		"### Fixed Bugs 🐛\n\n- Fix key ordering on document replacing by @noisersup in " +
		"https://api.github.com/repos/FerretDB/FerretDB/issues/1946\n" +
		"- Fix SASL response for `PLAIN` authentication by @b1ron in " +
		"https://api.github.com/repos/FerretDB/FerretDB/issues/1942\n" +
		"- Fix `$pop` operator error handling of non-existent path by @chilagrow in " +
		"https://api.github.com/repos/FerretDB/FerretDB/issues/1907\n\n" +
		"### Documentation 📄\n\n- Prepare v0.9.1 release by @AlekSi in " +
		"https://api.github.com/repos/FerretDB/FerretDB/issues/1958\n" +
		"- Fix broken link by @Fashander in https://api.github.com/repos/FerretDB/FerretDB/issues/1918\n" +
		"- Add blog post on \"MongoDB Alternatives: 5 Database Alternatives to MongoDB for 2023\"" +
		"by @Fashander in https://api.github.com/repos/FerretDB/FerretDB/issues/1911\n" +
		"- Bump deps by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1902\n\n" +
		"### Other Changes 🤖\n\n- Prepare v0.9.1 release by @AlekSi in " +
		"https://api.github.com/repos/FerretDB/FerretDB/issues/195+8\n" +
		"- Remove `skipTigrisPushdown` from tests by @noisersup in https://api.github.com/repos/FerretDB/FerretDB/issues/1957\n" +
		"- Rename function, add TODO by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1955\n" +
		"- Tweak CI settings by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1948\n" +
		"- Add `iterator.WithClose` helper by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1947\n" +
		"- Implement Tigris query iterator by @w84thesun in https://api.github.com/repos/FerretDB/FerretDB/issues/1924\n" +
		"- Remove unused parameter by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1919\n" +
		"- Bump Tigris by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1916\n" +
		"- Assorted internal tweaks by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1909\n" +
		"- Bump deps by @AlekSi in https://api.github.com/repos/FerretDB/FerretDB/issues/1902\n" +
		"- Use multiple Tigris instances to run tests by @chilagrow in " +
		"https://api.github.com/repos/FerretDB/FerretDB/issues/1878\n" +
		"- Add simple `otel` tracing to collect data from tests by @rumyantseva in " +
		"https://api.github.com/repos/FerretDB/FerretDB/issues/1863\n" +
		"- Rework on integration test setup by @chilagrow in https://api.github.com/repos/FerretDB/FerretDB/issues/1857\n\n"

	assert.Equal(t, expectedOutput, buf.String(), "Expected rendered markdown to be equal")
}
*/

func TestGenerateChangelogIntegration(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer r.Close()

	os.Stdout = w

	root, err := os.Getwd()
	require.NoError(t, err)

	root = filepath.Dir(filepath.Dir(root))

	run(root, "v1.21.0")

	w.Close()

	actual, err := io.ReadAll(r)
	require.NoError(t, err)

	date := time.Now().Format("2006-01-02")
	expected := "## [v1.21.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.21.0) (" + date + ")\n\n" +
		"### New Features 🎉\n\n" +
		"- Add experimental `SCRAM-SHA-1`/`SCRAM-SHA-256` authentication support by @henvic in " +
		"https://github.com/FerretDB/FerretDB/pull/4078\n\n" +
		"### Fixed Bugs 🐛\n\n" +
		"- Reorganize and fix `update`/`upsert` logic by @wazir-ahmed in " +
		"https://github.com/FerretDB/FerretDB/pull/4069\n\n" +
		"### Enhancements 🛠\n\n" +
		"- Improve capped collection cleanup by @wazir-ahmed in " +
		"https://github.com/FerretDB/FerretDB/pull/4118\n" +
		"- Make batch sizes configurable by @kropidlowsky in " +
		"https://github.com/FerretDB/FerretDB/pull/4149\n\n" +
		"### Documentation 📄\n\n" +
		"- Fix Codapi file error by @Fashander in " +
		"https://github.com/FerretDB/FerretDB/pull/4077\n" +
		"- Add Tembo QA blog post by @Fashander in " +
		"https://github.com/FerretDB/FerretDB/pull/4081\n" +
		"- Update correct image link by @Fashander in " +
		"https://github.com/FerretDB/FerretDB/pull/4116\n" +
		"- Add Pulumi blog post by @Fashander in " +
		"https://github.com/FerretDB/FerretDB/pull/4102\n" +
		"- Add Tembo to README by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4168\n" +
		"- Remove some closed issues from documentation by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4172\n\n" +
		"### Other Changes 🤖\n\n" +
		"- Use Go 1.22 and bump deps by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4094\n" +
		"- Add more fields to requests and responses by @rumyantseva in " +
		"https://github.com/FerretDB/FerretDB/pull/4096\n" +
		"- Revert SQLite version bump by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4106\n" +
		"- Refactor `bson2` package by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4105\n" +
		"- Use `bson2` package for wire queries and replies by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4108\n" +
		"- Make logger configurable in the embedded `ferretdb` package by @fadyat in " +
		"https://github.com/FerretDB/FerretDB/pull/4028\n" +
		"- Fix `envtool run test` `-run` and `-skip` flags by @henvic in " +
		"https://github.com/FerretDB/FerretDB/pull/4101\n" +
		"- Add MySQL backend collection by @adetunjii in " +
		"https://github.com/FerretDB/FerretDB/pull/4083\n" +
		"- Ignore `maxTimeMS` argument in `count`, `insert`, `update`, `delete` by @farit2000 in " +
		"https://github.com/FerretDB/FerretDB/pull/4121\n" +
		"- Use correct salt length by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4126\n" +
		"- Skip stuck tailable cursor test by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4131\n" +
		"- Enforce new authentication by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4075\n" +
		"- Replace `bson` with `bson2` in `wire` by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4110\n" +
		"- Improve `OP_MSG` validity checks by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4135\n" +
		"- Support speculative authenticate by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4111\n" +
		"- Add MySQL backend by @adetunjii in " +
		"https://github.com/FerretDB/FerretDB/pull/4137\n" +
		"- Fix `saslContinue` crashing due to not found authentication conversation by @henvic in " +
		"https://github.com/FerretDB/FerretDB/pull/4129\n" +
		"- Cleanup TODO for speculative authenticate by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4143\n" +
		"- Fix MySQL collection stats by @adetunjii in " +
		"https://github.com/FerretDB/FerretDB/pull/4145\n" +
		"- Use Go 1.22.1 by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4155\n" +
		"- Advertise SCRAM / SASL support in addition to PLAIN by @henvic in " +
		"https://github.com/FerretDB/FerretDB/pull/4113\n" +
		"- Add linter to check truncate tag in blog posts by @sbshah97 in " +
		"https://github.com/FerretDB/FerretDB/pull/4139\n" +
		"- Fix PLAIN mechanism authentication incorrectly working by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4163\n" +
		"- Improve `bson2` and `wire` logging by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4148\n" +
		"- Fix logging of deeply nested documents by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4167\n" +
		"- Support localhost exception by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4156\n" +
		"- Do not use the flow style in the diff output by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4170\n" +
		"- Do not use `fjson` by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4175\n" +
		"- Remove `fjson` package by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4176\n" +
		"- Fix `speculativeAuthenticate` panic on empty database by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4178\n" +
		"- Move old `bson` package by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4177\n" +
		"- Rename `bson2` to `bson` by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4179\n" +
		"- Move Docker build files by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4180\n" +
		"- Bump protobuf dependency to make CI happy by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4187\n" +
		"- Use authentication enabled docker for integration test by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4160\n" +
		"- Bump `pgx` by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4190\n\n" +
		"### New Contributors\n\n" +
		"- @farit2000 made their first contribution in " +
		"https://github.com/FerretDB/FerretDB/pull/4121\n" +
		"- @sbshah97 made their first contribution in " +
		"https://github.com/FerretDB/FerretDB/pull/4139\n\n" +
		"[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/63?closed=1).\n" +
		"[All commits](https://github.com/FerretDB/FerretDB/compare/v1.20.1...v1.21.0).\n"

	assert.Equal(t, expected, string(actual))
}
