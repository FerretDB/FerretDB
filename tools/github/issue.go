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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

// IssueStatus returns issue status.
// It uses cache.
//
// Returned error is something fatal.
// On rate limit, the error is logged once and (issueOpen, nil) is returned.
func (c *Client) IssueStatus(ctx context.Context, url, repo string, num int) (issueStatus, error) {
	start := time.Now()

	cache := &cacheFile{
		Issues: make(map[string]issue),
	}
	cacheRes := "miss"

	var res issueStatus

	// fast path without any locks

	data, err := os.ReadFile(c.cacheFilePath)
	if err == nil {
		_ = json.Unmarshal(data, cache)
		res = cache.Issues[url].Status
	}

	if res != "" {
		cacheRes = "fast hit"
	} else {
		// slow path

		noUpdate := fmt.Errorf("no need to update the cache file")

		err = lockedfile.Transform(c.cacheFilePath, func(data []byte) ([]byte, error) {
			cache.Issues = make(map[string]issue)

			if len(data) != 0 {
				if err = json.Unmarshal(data, cache); err != nil {
					return nil, err
				}
			}

			if res = cache.Issues[url].Status; res != "" {
				return nil, noUpdate
			}

			if res, err = c.checkIssueStatus(ctx, repo, num); err != nil {
				var rle *github.RateLimitError
				if !errors.As(err, &rle) {
					return nil, fmt.Errorf("%s: %s", url, err)
				}

				if cache.RateLimitReached {
					c.clientDebugF("Rate limit already reached: %s", url)
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
				cache.Issues[url] = issue{
					RefreshedAt: time.Now(),
					Status:      res,
				}
			}

			return json.MarshalIndent(cache, "", "  ")
		})

		if errors.Is(err, noUpdate) {
			cacheRes = "slow hit"
			err = nil
		}
	}

	c.cacheDebugF("%s: %s (%dms, %s)", url, res, time.Since(start).Milliseconds(), cacheRes)

	// when rate limited
	if err == nil && res == "" {
		res = issueOpen
	}

	return res, err
}

// checkIssueStatus checks issue status via GitHub API.
// It does not use cache.
func (c *Client) checkIssueStatus(ctx context.Context, repo string, num int) (issueStatus, error) {
	issue, resp, err := c.c.Issues.Get(ctx, "FerretDB", repo, num)
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

// Validate checks the issues status and returns its message (if it is not open).
func (status issueStatus) Validate(issueReference, url string) string {
	switch status {
	case issueOpen:
		// nothing
	case issueClosed:
		return fmt.Sprintf("invalid %s linked issue %s is closed", issueReference, url)
	case issueNotFound:
		return fmt.Sprintf("invalid %s linked issue %s is not found", issueReference, url)
	default:
		log.Panicf("unknown issue status: %s", status)
	}

	return ""
}
