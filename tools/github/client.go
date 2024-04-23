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

package github

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/FerretDB/gh"
	"github.com/google/go-github/v57/github"
)

// cacheFile stores information regarding rate limiting and the status of issues.
type cacheFile struct {
	Issues           map[string]issue `json:"issues"`
	RateLimitReached bool             `json:"rateLimitReached"`
}

// Client represent GitHub API Client with shared file cache.
type Client struct {
	c             *github.Client
	cacheFilePath string
	logf          gh.Printf
	cacheDebugF   gh.Printf
	clientDebugF  gh.Printf
	token         string
}

// NewClient creates a new client for the given cache file path and logging functions.
func NewClient(cacheFilePath string, logf, cacheDebugF, clientDebugf gh.Printf) (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")

	if logf == nil {
		panic("logf must be set")
	}

	if cacheDebugF == nil {
		panic("vf must be set")
	}

	if clientDebugf == nil {
		panic("debugf must be set")
	}

	c, err := gh.NewRESTClient(token, clientDebugf)
	if err != nil {
		return nil, err
	}

	return &Client{
		c:             c,
		cacheFilePath: cacheFilePath,
		token:         token,
		logf:          logf,
		cacheDebugF:   cacheDebugF,
		clientDebugF:  clientDebugf,
	}, nil
}

// CacheFilePath returns the path to the cache file.
func CacheFilePath(toolName string) (string, error) {
	// This tool is called for multiple packages in parallel,
	// with the current working directory set to the package directory.
	// To use the same cache file path, we first locate the root of the project by the .git directory.

	dir, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}

	for {
		fi, err := os.Stat(filepath.Join(dir, ".git"))
		if err == nil {
			if !fi.IsDir() {
				return "", fmt.Errorf(".git is not a directory")
			}

			break
		}

		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}

		dir = filepath.Dir(dir)
	}

	return filepath.Join(dir, "tmp", toolName, "cache.json"), nil
}
