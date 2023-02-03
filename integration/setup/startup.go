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

// Package setup provides integration tests setup helpers.
package setup

import (
	"strings"
	"sync/atomic"
)

// startupInitializer keeps tracks of the number of times
// tigris urls have been requested.
// See docker-compose.yml.
type startupInitializer struct {
	numCalls   *uint64
	tigrisURLs []string
}

// startupInitializer creates an instance of startupInitializer.
func newStartupInitializer() *startupInitializer {
	numCalls := uint64(0)
	urls := strings.Split(*tigrisURLsF, ",")

	return &startupInitializer{
		numCalls:   &numCalls,
		tigrisURLs: urls,
	}
}

// getNextTigrisURL gets the next URL of Tigris to be used
// for testing in Round Robin fashion.
func (p *startupInitializer) getNextTigrisURL() string {
	i := atomic.AddUint64(p.numCalls, 1)
	numUrls := uint64(len(p.tigrisURLs))

	return p.tigrisURLs[i%numUrls]
}
