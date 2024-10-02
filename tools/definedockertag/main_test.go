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

func TestDefineFerretDB(t *testing.T) {
	t.Run("pull_request", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_HEAD_REF":   "define-docker-tag",
			"GITHUB_REF_NAME":   "1/merge",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/all-in-one:pr-define-docker-tag",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/ferretdb-dev:pr-define-docker-tag",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_HEAD_REF":   "define-docker-tag",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/all-in-one:pr-define-docker-tag",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/ferretdb-dev:pr-define-docker-tag",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request/dependabot", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_HEAD_REF":   "dependabot/submodules/tests/mongo-go-driver-29d768e",
			"GITHUB_REF_NAME":   "58/merge",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/all-in-one:pr-mongo-go-driver-29d768e",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/ferretdb-dev:pr-mongo-go-driver-29d768e",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/main", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
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
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/release", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "releases/2.1",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
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
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/beta", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "v2.1.0-beta",
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ferretdb/all-in-one:2.1.0-beta",
				"ghcr.io/ferretdb/all-in-one:2.1.0-beta",
				"quay.io/ferretdb/all-in-one:2.1.0-beta",
			},
			developmentImages: []string{
				"ferretdb/ferretdb-dev:2.1.0-beta",
				"ghcr.io/ferretdb/ferretdb-dev:2.1.0-beta",
				"quay.io/ferretdb/ferretdb-dev:2.1.0-beta",
			},
			productionImages: []string{
				"ferretdb/ferretdb:2.1.0-beta",
				"ghcr.io/ferretdb/ferretdb:2.1.0-beta",
				"quay.io/ferretdb/ferretdb:2.1.0-beta",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/release", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "v2.1.0",
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
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
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/wrong", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "2.1.0", // no leading v
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		_, err := define(getenv)
		require.Error(t, err)
	})

	t.Run("schedule", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "schedule",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
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
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("workflow_run", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "workflow_run",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
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
		}
		assert.Equal(t, expected, actual)
	})
}

func TestDefineOtherOrg(t *testing.T) {
	t.Run("pull_request", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_HEAD_REF":   "define-docker-tag",
			"GITHUB_REF_NAME":   "1/merge",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "OtherOrg/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/otherorg/all-in-one:pr-define-docker-tag",
			},
			developmentImages: []string{
				"ghcr.io/otherorg/ferretdb-dev:pr-define-docker-tag",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_HEAD_REF":   "define-docker-tag",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "OtherOrg/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/otherorg/all-in-one:pr-define-docker-tag",
			},
			developmentImages: []string{
				"ghcr.io/otherorg/ferretdb-dev:pr-define-docker-tag",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request/dependabot", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_HEAD_REF":   "dependabot/submodules/tests/mongo-go-driver-29d768e",
			"GITHUB_REF_NAME":   "58/merge",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "OtherOrg/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/otherorg/all-in-one:pr-mongo-go-driver-29d768e",
			},
			developmentImages: []string{
				"ghcr.io/otherorg/ferretdb-dev:pr-mongo-go-driver-29d768e",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/main", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "OtherOrg/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/otherorg/all-in-one:main",
			},
			developmentImages: []string{
				"ghcr.io/otherorg/ferretdb-dev:main",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/release", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "releases/2.1",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "OtherOrg/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/otherorg/all-in-one:releases-2.1",
			},
			developmentImages: []string{
				"ghcr.io/otherorg/ferretdb-dev:releases-2.1",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/beta", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "v2.1.0-beta",
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "OtherOrg/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/otherorg/all-in-one:2.1.0-beta",
			},
			developmentImages: []string{
				"ghcr.io/otherorg/ferretdb-dev:2.1.0-beta",
			},
			productionImages: []string{
				"ghcr.io/otherorg/ferretdb:2.1.0-beta",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/release", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "v2.1.0",
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "OtherOrg/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/otherorg/all-in-one:2",
				"ghcr.io/otherorg/all-in-one:2.1",
				"ghcr.io/otherorg/all-in-one:2.1.0",
			},
			developmentImages: []string{
				"ghcr.io/otherorg/ferretdb-dev:2",
				"ghcr.io/otherorg/ferretdb-dev:2.1",
				"ghcr.io/otherorg/ferretdb-dev:2.1.0",
			},
			productionImages: []string{
				"ghcr.io/otherorg/ferretdb:2",
				"ghcr.io/otherorg/ferretdb:2.1",
				"ghcr.io/otherorg/ferretdb:2.1.0",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/wrong", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "2.1.0", // no leading v
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "OtherOrg/FerretDB",
		})

		_, err := define(getenv)
		require.Error(t, err)
	})

	t.Run("schedule", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "schedule",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "OtherOrg/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/otherorg/all-in-one:main",
			},
			developmentImages: []string{
				"ghcr.io/otherorg/ferretdb-dev:main",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("workflow_run", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "workflow_run",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "OtherOrg/FerretDB",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/otherorg/all-in-one:main",
			},
			developmentImages: []string{
				"ghcr.io/otherorg/ferretdb-dev:main",
			},
		}
		assert.Equal(t, expected, actual)
	})
}

func TestDefineOtherRepo(t *testing.T) {
	t.Run("pull_request", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_HEAD_REF":   "define-docker-tag",
			"GITHUB_REF_NAME":   "1/merge",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/OtherRepo",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/otherrepo-all-in-one:pr-define-docker-tag",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/otherrepo-dev:pr-define-docker-tag",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request_target", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request_target",
			"GITHUB_HEAD_REF":   "define-docker-tag",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/OtherRepo",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/otherrepo-all-in-one:pr-define-docker-tag",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/otherrepo-dev:pr-define-docker-tag",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("pull_request/dependabot", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "main",
			"GITHUB_EVENT_NAME": "pull_request",
			"GITHUB_HEAD_REF":   "dependabot/submodules/tests/mongo-go-driver-29d768e",
			"GITHUB_REF_NAME":   "58/merge",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/OtherRepo",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/otherrepo-all-in-one:pr-mongo-go-driver-29d768e",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/otherrepo-dev:pr-mongo-go-driver-29d768e",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/main", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/OtherRepo",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/otherrepo-all-in-one:main",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/otherrepo-dev:main",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/release", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "releases/2.1",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/OtherRepo",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/otherrepo-all-in-one:releases-2.1",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/otherrepo-dev:releases-2.1",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/beta", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "v2.1.0-beta",
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "FerretDB/OtherRepo",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/otherrepo-all-in-one:2.1.0-beta",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/otherrepo-dev:2.1.0-beta",
			},
			productionImages: []string{
				"ghcr.io/ferretdb/otherrepo:2.1.0-beta",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/release", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "v2.1.0",
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "FerretDB/OtherRepo",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/otherrepo-all-in-one:2",
				"ghcr.io/ferretdb/otherrepo-all-in-one:2.1",
				"ghcr.io/ferretdb/otherrepo-all-in-one:2.1.0",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/otherrepo-dev:2",
				"ghcr.io/ferretdb/otherrepo-dev:2.1",
				"ghcr.io/ferretdb/otherrepo-dev:2.1.0",
			},
			productionImages: []string{
				"ghcr.io/ferretdb/otherrepo:2",
				"ghcr.io/ferretdb/otherrepo:2.1",
				"ghcr.io/ferretdb/otherrepo:2.1.0",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("push/tag/wrong", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "push",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "2.1.0", // no leading v
			"GITHUB_REF_TYPE":   "tag",
			"GITHUB_REPOSITORY": "FerretDB/OtherRepo",
		})

		_, err := define(getenv)
		require.Error(t, err)
	})

	t.Run("schedule", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "schedule",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/OtherRepo",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/otherrepo-all-in-one:main",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/otherrepo-dev:main",
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("workflow_run", func(t *testing.T) {
		getenv := getEnvFunc(t, map[string]string{
			"GITHUB_BASE_REF":   "",
			"GITHUB_EVENT_NAME": "workflow_run",
			"GITHUB_HEAD_REF":   "",
			"GITHUB_REF_NAME":   "main",
			"GITHUB_REF_TYPE":   "branch",
			"GITHUB_REPOSITORY": "FerretDB/OtherRepo",
		})

		actual, err := define(getenv)
		require.NoError(t, err)

		expected := &result{
			allInOneImages: []string{
				"ghcr.io/ferretdb/otherrepo-all-in-one:main",
			},
			developmentImages: []string{
				"ghcr.io/ferretdb/otherrepo-dev:main",
			},
		}
		assert.Equal(t, expected, actual)
	})
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
