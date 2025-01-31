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
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/FerretDB/gh"
	"github.com/google/go-github/v66/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMilestone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	ctx := context.Background()
	client, err := gh.NewRESTClient(os.Getenv("GITHUB_TOKEN"), nil)
	require.NoError(t, err)

	milestoneTitle := "v0.9.1"

	milestone, err := getMilestone(ctx, client, milestoneTitle)
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
}

func TestGenerateChangelog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	root, err := os.Getwd()
	require.NoError(t, err)

	root = filepath.Dir(filepath.Dir(root))

	var buf bytes.Buffer
	run(&buf, root, "v1.21.0", "v1.20.1")

	date := time.Now().Format("2006-01-02")
	expected := "\n" +
		"## [v1.21.0](https://github.com/FerretDB/FerretDB/releases/tag/v1.21.0) (" + date + ")\n\n" +
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
		"- Add Pulumi blog post by @Fashander in " +
		"https://github.com/FerretDB/FerretDB/pull/4102\n" +
		"- Update correct image link by @Fashander in " +
		"https://github.com/FerretDB/FerretDB/pull/4116\n" +
		"- Add Tembo to README by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4168\n" +
		"- Remove some closed issues from documentation by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4172\n\n" +
		"### Other Changes 🤖\n\n" +
		"- Make logger configurable in the embedded `ferretdb` package by @fadyat in " +
		"https://github.com/FerretDB/FerretDB/pull/4028\n" +
		"- Enforce new authentication by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4075\n" +
		"- Add MySQL backend collection by @adetunjii in " +
		"https://github.com/FerretDB/FerretDB/pull/4083\n" +
		"- Use Go 1.22 and bump deps by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4094\n" +
		"- Add more fields to requests and responses by @rumyantseva in " +
		"https://github.com/FerretDB/FerretDB/pull/4096\n" +
		"- Fix `envtool run test` `-run` and `-skip` flags by @henvic in " +
		"https://github.com/FerretDB/FerretDB/pull/4101\n" +
		"- Refactor `bson2` package by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4105\n" +
		"- Revert SQLite version bump by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4106\n" +
		"- Use `bson2` package for wire queries and replies by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4108\n" +
		"- Replace `bson` with `bson2` in `wire` by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4110\n" +
		"- Support speculative authenticate by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4111\n" +
		"- Advertise SCRAM / SASL support in addition to PLAIN by @henvic in " +
		"https://github.com/FerretDB/FerretDB/pull/4113\n" +
		"- Ignore `maxTimeMS` argument in `count`, `insert`, `update`, `delete` by @farit2000 in " +
		"https://github.com/FerretDB/FerretDB/pull/4121\n" +
		"- Use correct salt length by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4126\n" +
		"- Fix `saslContinue` crashing due to not found authentication conversation by @henvic in " +
		"https://github.com/FerretDB/FerretDB/pull/4129\n" +
		"- Skip stuck tailable cursor test by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4131\n" +
		"- Improve `OP_MSG` validity checks by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4135\n" +
		"- Add MySQL backend by @adetunjii in " +
		"https://github.com/FerretDB/FerretDB/pull/4137\n" +
		"- Add linter to check truncate tag in blog posts by @sbshah97 in " +
		"https://github.com/FerretDB/FerretDB/pull/4139\n" +
		"- Cleanup TODO for speculative authenticate by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4143\n" +
		"- Fix MySQL collection stats by @adetunjii in " +
		"https://github.com/FerretDB/FerretDB/pull/4145\n" +
		"- Improve `bson2` and `wire` logging by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4148\n" +
		"- Use Go 1.22.1 by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4155\n" +
		"- Support localhost exception by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4156\n" +
		"- Use authentication enabled docker for integration test by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4160\n" +
		"- Fix PLAIN mechanism authentication incorrectly working by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4163\n" +
		"- Fix logging of deeply nested documents by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4167\n" +
		"- Do not use the flow style in the diff output by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4170\n" +
		"- Do not use `fjson` by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4175\n" +
		"- Remove `fjson` package by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4176\n" +
		"- Move old `bson` package by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4177\n" +
		"- Fix `speculativeAuthenticate` panic on empty database by @chilagrow in " +
		"https://github.com/FerretDB/FerretDB/pull/4178\n" +
		"- Rename `bson2` to `bson` by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4179\n" +
		"- Move Docker build files by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4180\n" +
		"- Bump protobuf dependency to make CI happy by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4187\n" +
		"- Bump `pgx` by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4190\n" +
		"- Prepare v1.21.0 release by @AlekSi in " +
		"https://github.com/FerretDB/FerretDB/pull/4195\n\n" +
		"[All closed issues and pull requests](https://github.com/FerretDB/FerretDB/milestone/63?closed=1).\n" +
		"[All commits](https://github.com/FerretDB/FerretDB/compare/v1.20.1...v1.21.0).\n"

	assert.Equal(t, expected, buf.String())
}
