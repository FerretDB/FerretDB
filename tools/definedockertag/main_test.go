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
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getEnvFunc implements [os.Getenv] for testing.
func getEnvFunc(t *testing.T, env map[string]string) func(string) string {
	t.Helper()

	return func(key string) string {
		val, ok := env[key]
		require.True(t, ok, "missing key %q", key)

		return val
	}
}

type testCase struct {
	env      map[string]string
	expected *result
}

func TestDefine(t *testing.T) {
	for name, tc := range map[string]testCase{
		"pull_request": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "main",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "define-docker-tag",
				"GITHUB_REF_NAME":   "1/merge",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/ferretdb/all-in-one:pr-define-docker-tag",
				},
				developmentImages: []string{
					"ghcr.io/ferretdb/ferretdb-dev:pr-define-docker-tag",
				},
			},
		},
		"pull_request-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "main",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "define-docker-tag",
				"GITHUB_REF_NAME":   "1/merge",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:pr-define-docker-tag",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:pr-define-docker-tag",
				},
			},
		},

		"pull_request/dependabot": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "main",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "dependabot/submodules/tests/mongo-go-driver-29d768e",
				"GITHUB_REF_NAME":   "58/merge",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/ferretdb/all-in-one:pr-mongo-go-driver-29d768e",
				},
				developmentImages: []string{
					"ghcr.io/ferretdb/ferretdb-dev:pr-mongo-go-driver-29d768e",
				},
			},
		},
		"pull_request/dependabot-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "main",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "dependabot/submodules/tests/mongo-go-driver-29d768e",
				"GITHUB_REF_NAME":   "58/merge",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:pr-mongo-go-driver-29d768e",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:pr-mongo-go-driver-29d768e",
				},
			},
		},

		"pull_request_target": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "main",
				"GITHUB_EVENT_NAME": "pull_request_target",
				"GITHUB_HEAD_REF":   "define-docker-tag",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/ferretdb/all-in-one:pr-define-docker-tag",
				},
				developmentImages: []string{
					"ghcr.io/ferretdb/ferretdb-dev:pr-define-docker-tag",
				},
			},
		},
		"pull_request_target-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "main",
				"GITHUB_EVENT_NAME": "pull_request_target",
				"GITHUB_HEAD_REF":   "define-docker-tag",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:pr-define-docker-tag",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:pr-define-docker-tag",
				},
			},
		},

		"push/main": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				allInOneImages: []string{
					"ferretdb/all-in-one:main",
					"ghcr.io/ferretdb/all-in-one:main",
					"quay.io/ferretdb/all-in-one:main",
				},
				developmentImages: []string{
					"ferretdb/ferretdb-dev:main",
					"ghcr.io/ferretdb/ferretdb-dev:main",
					"quay.io/ferretdb/ferretdb-dev:main",
				},
			},
		},
		"push/main-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:main",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:main",
				},
			},
		},

		"push/main-v1": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main-v1",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				allInOneImages: []string{
					"ferretdb/all-in-one:main-v1",
					"ghcr.io/ferretdb/all-in-one:main-v1",
					"quay.io/ferretdb/all-in-one:main-v1",
				},
				developmentImages: []string{
					"ferretdb/ferretdb-dev:main-v1",
					"ghcr.io/ferretdb/ferretdb-dev:main-v1",
					"quay.io/ferretdb/ferretdb-dev:main-v1",
				},
			},
		},
		"push/main-v1-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main-v1",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:main-v1",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:main-v1",
				},
			},
		},

		"push/release": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "releases/2.1",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				allInOneImages: []string{
					"ferretdb/all-in-one:releases-2.1",
					"ghcr.io/ferretdb/all-in-one:releases-2.1",
					"quay.io/ferretdb/all-in-one:releases-2.1",
				},
				developmentImages: []string{
					"ferretdb/ferretdb-dev:releases-2.1",
					"ghcr.io/ferretdb/ferretdb-dev:releases-2.1",
					"quay.io/ferretdb/ferretdb-dev:releases-2.1",
				},
			},
		},
		"push/release-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "releases/2.1",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:releases-2.1",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:releases-2.1",
				},
			},
		},

		"push/tag/beta1": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v1.26.0-beta",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				allInOneImages: []string{
					"ferretdb/all-in-one:1.26.0-beta",
					"ghcr.io/ferretdb/all-in-one:1.26.0-beta",
					"quay.io/ferretdb/all-in-one:1.26.0-beta",
				},
				developmentImages: []string{
					"ferretdb/ferretdb-dev:1.26.0-beta",
					"ghcr.io/ferretdb/ferretdb-dev:1.26.0-beta",
					"quay.io/ferretdb/ferretdb-dev:1.26.0-beta",
				},
				productionImages: []string{
					"ferretdb/ferretdb:1.26.0-beta",
					"ghcr.io/ferretdb/ferretdb:1.26.0-beta",
					"quay.io/ferretdb/ferretdb:1.26.0-beta",
				},
			},
		},
		"push/tag/beta1-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v1.26.0-beta",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:1.26.0-beta",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:1.26.0-beta",
				},
				productionImages: []string{
					"ghcr.io/otherorg/otherrepo:1.26.0-beta",
				},
			},
		},

		"push/tag/beta2": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v2.1.0-beta",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				// add all tags except latest while v2 is not GA
				allInOneImages: []string{
					"ferretdb/all-in-one:2",
					"ferretdb/all-in-one:2.1",
					"ferretdb/all-in-one:2.1.0",
					"ferretdb/all-in-one:2.1.0-beta",
					"ghcr.io/ferretdb/all-in-one:2",
					"ghcr.io/ferretdb/all-in-one:2.1",
					"ghcr.io/ferretdb/all-in-one:2.1.0",
					"ghcr.io/ferretdb/all-in-one:2.1.0-beta",
					"quay.io/ferretdb/all-in-one:2",
					"quay.io/ferretdb/all-in-one:2.1",
					"quay.io/ferretdb/all-in-one:2.1.0",
					"quay.io/ferretdb/all-in-one:2.1.0-beta",
				},
				developmentImages: []string{
					"ferretdb/ferretdb-dev:2",
					"ferretdb/ferretdb-dev:2.1",
					"ferretdb/ferretdb-dev:2.1.0",
					"ferretdb/ferretdb-dev:2.1.0-beta",
					"ghcr.io/ferretdb/ferretdb-dev:2",
					"ghcr.io/ferretdb/ferretdb-dev:2.1",
					"ghcr.io/ferretdb/ferretdb-dev:2.1.0",
					"ghcr.io/ferretdb/ferretdb-dev:2.1.0-beta",
					"quay.io/ferretdb/ferretdb-dev:2",
					"quay.io/ferretdb/ferretdb-dev:2.1",
					"quay.io/ferretdb/ferretdb-dev:2.1.0",
					"quay.io/ferretdb/ferretdb-dev:2.1.0-beta",
				},
				productionImages: []string{
					"ferretdb/ferretdb:2",
					"ferretdb/ferretdb:2.1",
					"ferretdb/ferretdb:2.1.0",
					"ferretdb/ferretdb:2.1.0-beta",
					"ghcr.io/ferretdb/ferretdb:2",
					"ghcr.io/ferretdb/ferretdb:2.1",
					"ghcr.io/ferretdb/ferretdb:2.1.0",
					"ghcr.io/ferretdb/ferretdb:2.1.0-beta",
					"quay.io/ferretdb/ferretdb:2",
					"quay.io/ferretdb/ferretdb:2.1",
					"quay.io/ferretdb/ferretdb:2.1.0",
					"quay.io/ferretdb/ferretdb:2.1.0-beta",
				},
			},
		},
		"push/tag/beta2-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v2.1.0-beta",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				// add all tags except latest while v2 is not GA
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:2",
					"ghcr.io/otherorg/otherrepo-all-in-one:2.1",
					"ghcr.io/otherorg/otherrepo-all-in-one:2.1.0",
					"ghcr.io/otherorg/otherrepo-all-in-one:2.1.0-beta",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:2",
					"ghcr.io/otherorg/otherrepo-dev:2.1",
					"ghcr.io/otherorg/otherrepo-dev:2.1.0",
					"ghcr.io/otherorg/otherrepo-dev:2.1.0-beta",
				},
				productionImages: []string{
					"ghcr.io/otherorg/otherrepo:2",
					"ghcr.io/otherorg/otherrepo:2.1",
					"ghcr.io/otherorg/otherrepo:2.1.0",
					"ghcr.io/otherorg/otherrepo:2.1.0-beta",
				},
			},
		},

		"push/tag/release1": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v1.26.0",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				// latest is v1 for now
				allInOneImages: []string{
					"ferretdb/all-in-one:1",
					"ferretdb/all-in-one:1.26",
					"ferretdb/all-in-one:1.26.0",
					"ferretdb/all-in-one:latest",
					"ghcr.io/ferretdb/all-in-one:1",
					"ghcr.io/ferretdb/all-in-one:1.26",
					"ghcr.io/ferretdb/all-in-one:1.26.0",
					"ghcr.io/ferretdb/all-in-one:latest",
					"quay.io/ferretdb/all-in-one:1",
					"quay.io/ferretdb/all-in-one:1.26",
					"quay.io/ferretdb/all-in-one:1.26.0",
					"quay.io/ferretdb/all-in-one:latest",
				},
				developmentImages: []string{
					"ferretdb/ferretdb-dev:1",
					"ferretdb/ferretdb-dev:1.26",
					"ferretdb/ferretdb-dev:1.26.0",
					"ferretdb/ferretdb-dev:latest",
					"ghcr.io/ferretdb/ferretdb-dev:1",
					"ghcr.io/ferretdb/ferretdb-dev:1.26",
					"ghcr.io/ferretdb/ferretdb-dev:1.26.0",
					"ghcr.io/ferretdb/ferretdb-dev:latest",
					"quay.io/ferretdb/ferretdb-dev:1",
					"quay.io/ferretdb/ferretdb-dev:1.26",
					"quay.io/ferretdb/ferretdb-dev:1.26.0",
					"quay.io/ferretdb/ferretdb-dev:latest",
				},
				productionImages: []string{
					"ferretdb/ferretdb:1",
					"ferretdb/ferretdb:1.26",
					"ferretdb/ferretdb:1.26.0",
					"ferretdb/ferretdb:latest",
					"ghcr.io/ferretdb/ferretdb:1",
					"ghcr.io/ferretdb/ferretdb:1.26",
					"ghcr.io/ferretdb/ferretdb:1.26.0",
					"ghcr.io/ferretdb/ferretdb:latest",
					"quay.io/ferretdb/ferretdb:1",
					"quay.io/ferretdb/ferretdb:1.26",
					"quay.io/ferretdb/ferretdb:1.26.0",
					"quay.io/ferretdb/ferretdb:latest",
				},
			},
		},
		"push/tag/release1-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v1.26.0",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				// latest is v1 for now
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:1",
					"ghcr.io/otherorg/otherrepo-all-in-one:1.26",
					"ghcr.io/otherorg/otherrepo-all-in-one:1.26.0",
					"ghcr.io/otherorg/otherrepo-all-in-one:latest",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:1",
					"ghcr.io/otherorg/otherrepo-dev:1.26",
					"ghcr.io/otherorg/otherrepo-dev:1.26.0",
					"ghcr.io/otherorg/otherrepo-dev:latest",
				},
				productionImages: []string{
					"ghcr.io/otherorg/otherrepo:1",
					"ghcr.io/otherorg/otherrepo:1.26",
					"ghcr.io/otherorg/otherrepo:1.26.0",
					"ghcr.io/otherorg/otherrepo:latest",
				},
			},
		},

		"push/tag/release2": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v2.1.0",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				// latest is v1, not v2, for now
				allInOneImages: []string{
					"ferretdb/all-in-one:2",
					"ferretdb/all-in-one:2.1",
					"ferretdb/all-in-one:2.1.0",
					"ghcr.io/ferretdb/all-in-one:2",
					"ghcr.io/ferretdb/all-in-one:2.1",
					"ghcr.io/ferretdb/all-in-one:2.1.0",
					"quay.io/ferretdb/all-in-one:2",
					"quay.io/ferretdb/all-in-one:2.1",
					"quay.io/ferretdb/all-in-one:2.1.0",
				},
				developmentImages: []string{
					"ferretdb/ferretdb-dev:2",
					"ferretdb/ferretdb-dev:2.1",
					"ferretdb/ferretdb-dev:2.1.0",
					"ghcr.io/ferretdb/ferretdb-dev:2",
					"ghcr.io/ferretdb/ferretdb-dev:2.1",
					"ghcr.io/ferretdb/ferretdb-dev:2.1.0",
					"quay.io/ferretdb/ferretdb-dev:2",
					"quay.io/ferretdb/ferretdb-dev:2.1",
					"quay.io/ferretdb/ferretdb-dev:2.1.0",
				},
				productionImages: []string{
					"ferretdb/ferretdb:2",
					"ferretdb/ferretdb:2.1",
					"ferretdb/ferretdb:2.1.0",
					"ghcr.io/ferretdb/ferretdb:2",
					"ghcr.io/ferretdb/ferretdb:2.1",
					"ghcr.io/ferretdb/ferretdb:2.1.0",
					"quay.io/ferretdb/ferretdb:2",
					"quay.io/ferretdb/ferretdb:2.1",
					"quay.io/ferretdb/ferretdb:2.1.0",
				},
			},
		},
		"push/tag/release2-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "v2.1.0",
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				// latest is v1, not v2, for now
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:2",
					"ghcr.io/otherorg/otherrepo-all-in-one:2.1",
					"ghcr.io/otherorg/otherrepo-all-in-one:2.1.0",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:2",
					"ghcr.io/otherorg/otherrepo-dev:2.1",
					"ghcr.io/otherorg/otherrepo-dev:2.1.0",
				},
				productionImages: []string{
					"ghcr.io/otherorg/otherrepo:2",
					"ghcr.io/otherorg/otherrepo:2.1",
					"ghcr.io/otherorg/otherrepo:2.1.0",
				},
			},
		},

		"push/tag/wrong": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "2.1.0", // no leading v
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
		},
		"push/tag/wrong-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "2.1.0", // no leading v
				"GITHUB_REF_TYPE":   "tag",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
		},

		"schedule": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "schedule",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				allInOneImages: []string{
					"ferretdb/all-in-one:main",
					"ghcr.io/ferretdb/all-in-one:main",
					"quay.io/ferretdb/all-in-one:main",
				},
				developmentImages: []string{
					"ferretdb/ferretdb-dev:main",
					"ghcr.io/ferretdb/ferretdb-dev:main",
					"quay.io/ferretdb/ferretdb-dev:main",
				},
			},
		},
		"schedule-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "schedule",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:main",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:main",
				},
			},
		},

		"workflow_run": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "workflow_run",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "FerretDB/FerretDB",
			},
			expected: &result{
				allInOneImages: []string{
					"ferretdb/all-in-one:main",
					"ghcr.io/ferretdb/all-in-one:main",
					"quay.io/ferretdb/all-in-one:main",
				},
				developmentImages: []string{
					"ferretdb/ferretdb-dev:main",
					"ghcr.io/ferretdb/ferretdb-dev:main",
					"quay.io/ferretdb/ferretdb-dev:main",
				},
			},
		},
		"workflow_run-other": {
			env: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "workflow_run",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF_NAME":   "main",
				"GITHUB_REF_TYPE":   "branch",
				"GITHUB_REPOSITORY": "OtherOrg/OtherRepo",
			},
			expected: &result{
				allInOneImages: []string{
					"ghcr.io/otherorg/otherrepo-all-in-one:main",
				},
				developmentImages: []string{
					"ghcr.io/otherorg/otherrepo-dev:main",
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			actual, err := define(getEnvFunc(t, tc.env))
			if tc.expected == nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

//nolint:lll // it is more readable this way
func TestImageURL(t *testing.T) {
	// expected URLs should work
	assert.Equal(t, "https://ghcr.io/ferretdb/all-in-one:pr-define-docker-tag", imageURL("ghcr.io/ferretdb/all-in-one:pr-define-docker-tag"))
	assert.Equal(t, "https://quay.io/ferretdb/all-in-one:pr-define-docker-tag", imageURL("quay.io/ferretdb/all-in-one:pr-define-docker-tag"))
	assert.Equal(t, "https://hub.docker.com/r/ferretdb/all-in-one/tags", imageURL("ferretdb/all-in-one:pr-define-docker-tag"))
}

func TestResults(t *testing.T) {
	dir := t.TempDir()

	summaryF, err := os.CreateTemp(dir, "summary")
	require.NoError(t, err)
	defer summaryF.Close() //nolint:errcheck // temporary file for testing

	outputF, err := os.CreateTemp(dir, "output")
	require.NoError(t, err)
	defer outputF.Close() //nolint:errcheck // temporary file for testing

	var stdout bytes.Buffer
	getenv := getEnvFunc(t, map[string]string{
		"GITHUB_STEP_SUMMARY": summaryF.Name(),
		"GITHUB_OUTPUT":       outputF.Name(),
	})
	action := githubactions.New(githubactions.WithGetenv(getenv), githubactions.WithWriter(&stdout))

	result := &result{
		allInOneImages: []string{
			"ferretdb/all-in-one:2.1.0",
		},
		developmentImages: []string{
			"ghcr.io/ferretdb/ferretdb-dev:2",
		},
		productionImages: []string{
			"quay.io/ferretdb/ferretdb:latest",
		},
	}

	setResults(action, result)

	expectedStdout := strings.ReplaceAll(`
 |Type        |Image                                                                            |
 |----        |-----                                                                            |
 |All-in-one  |['ferretdb/all-in-one:2.1.0'](https://hub.docker.com/r/ferretdb/all-in-one/tags) |
 |Development |['ghcr.io/ferretdb/ferretdb-dev:2'](https://ghcr.io/ferretdb/ferretdb-dev:2)     |
 |Production  |['quay.io/ferretdb/ferretdb:latest'](https://quay.io/ferretdb/ferretdb:latest)   |

`[1:], "'", "`")
	assert.Equal(t, expectedStdout, stdout.String(), "stdout does not match")

	expectedSummary := strings.ReplaceAll(`
 |Type        |Image                                                                            |
 |----        |-----                                                                            |
 |All-in-one  |['ferretdb/all-in-one:2.1.0'](https://hub.docker.com/r/ferretdb/all-in-one/tags) |
 |Development |['ghcr.io/ferretdb/ferretdb-dev:2'](https://ghcr.io/ferretdb/ferretdb-dev:2)     |
 |Production  |['quay.io/ferretdb/ferretdb:latest'](https://quay.io/ferretdb/ferretdb:latest)   |

`[1:], "'", "`")
	b, err := io.ReadAll(summaryF)
	require.NoError(t, err)
	assert.Equal(t, expectedSummary, string(b), "summary does not match")

	expectedOutput := `
all_in_one_images<<_GitHubActionsFileCommandDelimeter_
ferretdb/all-in-one:2.1.0
_GitHubActionsFileCommandDelimeter_
development_images<<_GitHubActionsFileCommandDelimeter_
ghcr.io/ferretdb/ferretdb-dev:2
_GitHubActionsFileCommandDelimeter_
production_images<<_GitHubActionsFileCommandDelimeter_
quay.io/ferretdb/ferretdb:latest
_GitHubActionsFileCommandDelimeter_
`[1:]
	b, err = io.ReadAll(outputF)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, string(b), "output parameters does not match")
}
