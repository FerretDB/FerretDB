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
	"fmt"
	"log"
	"time"
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
