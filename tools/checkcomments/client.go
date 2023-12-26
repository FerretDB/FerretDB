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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/FerretDB/gh"
	"github.com/google/go-github/v57/github"
	"github.com/rogpeppe/go-internal/lockedfile"
)

// issueStatus represents a known issue status.
type issueStatus string

// Known issue statuses.
const (
	issueOpen     issueStatus = "open"
	issueClosed   issueStatus = "closed"
	issueNotFound issueStatus = "not found"
)

// issue represents a single cached issue.
type issue struct {
	RefreshedAt time.Time   `json:"refreshedAt"`
	Status      issueStatus `json:"status"`
}

// cacheFile stores information regarding rate limiting and the status of issues.
type cacheFile struct {
	Issues           map[string]issue `json:"issues"`
	RateLimitReached bool             `json:"rateLimitReached"`
}

// client represent GitHub API client with shared file cache.
type client struct {
	c             *github.Client
	cacheFilePath string
	logf          gh.Printf
	debugf        gh.Printf
	token         string
}

// newClient creates a new client for the given cache file path and logging functions.
func newClient(cacheFilePath string, logf, debugf gh.Printf) (*client, error) {
	token := os.Getenv("GITHUB_TOKEN")

	if logf == nil {
		panic("logf must be set")
	}

	if debugf == nil {
		panic("debugf must be set")
	}

	c, err := gh.NewRESTClient(token, debugf)
	if err != nil {
		return nil, err
	}

	return &client{
		c:             c,
		cacheFilePath: cacheFilePath,
		token:         token,
		logf:          logf,
		debugf:        debugf,
	}, nil
}

// IssueStatus returns issue status.
// It uses cache.
//
// Returned error is something fatal.
// On rate limit, the error is logged once and (issueOpen, nil) is returned.
func (c *client) IssueStatus(ctx context.Context, num int) (issueStatus, error) {
	var res issueStatus
	noUpdate := fmt.Errorf("no need to update the cache file")

	err := lockedfile.Transform(c.cacheFilePath, func(data []byte) ([]byte, error) {
		cache := &cacheFile{
			Issues: make(map[string]issue),
		}

		if len(data) != 0 {
			if err := json.Unmarshal(data, cache); err != nil {
				return nil, err
			}
		}

		url := fmt.Sprintf("https://github.com/FerretDB/FerretDB/issues/%d", num)

		if res = cache.Issues[url].Status; res != "" {
			c.debugf("Cache hit for %s: %s", url, res)
			return nil, noUpdate
		}

		var err error
		if res, err = c.checkIssueStatus(ctx, num); err != nil {
			var rle *github.RateLimitError
			if !errors.As(err, &rle) {
				return nil, fmt.Errorf("%s: %s", url, err)
			}

			if cache.RateLimitReached {
				c.debugf("Rate limit already reached: %s", url)
				return nil, noUpdate
			}

			cache.RateLimitReached = true

			msg := "Rate limit reached: " + err.Error()
			if c.token == "" {
				msg += "\nPlease set a GITHUB_TOKEN as described at " +
					"https://github.com/FerretDB/FerretDB/blob/main/CONTRIBUTING.md#setting-a-github_token"
			}
			c.logf("%s", msg)
		}

		// unless rate limited
		if res != "" {
			c.debugf("Cache miss for %s: %s", url, res)
			cache.Issues[url] = issue{
				RefreshedAt: time.Now(),
				Status:      res,
			}
		}

		return json.MarshalIndent(cache, "", "  ")
	})

	if errors.Is(err, noUpdate) {
		err = nil
	}

	// when rate limited
	if err == nil && res == "" {
		res = issueOpen
	}

	return res, err
}

// checkIssueStatus checks issue status via GitHub API.
// It does not use cache.
func (c *client) checkIssueStatus(ctx context.Context, num int) (issueStatus, error) {
	issue, resp, err := c.c.Issues.Get(ctx, "FerretDB", "FerretDB", num)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return issueNotFound, nil
		}

		return "", err
	}

	switch s := issue.GetState(); s {
	case "open":
		return issueOpen, nil
	case "closed":
		return issueClosed, nil
	default:
		return "", fmt.Errorf("unknown issue state: %q", s)
	}
}

// cacheFilePath returns the path to the cache file.
func cacheFilePath() (string, error) {
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

	return filepath.Join(dir, "tmp", "checkcomments", "cache.json"), nil
}
