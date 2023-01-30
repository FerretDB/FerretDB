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
	"sync/atomic"
)

// ports are available port of Tigris.
var ports = []uint16{8081, 8082, 8083, 8084, 8085}

// startupInitializer keeps tracks of the number of times
// ports have been requested.
type startupInitializer struct {
	nPortCalls *uint64
}

// startupInitializer creates an instance of startupInitializer.
func newStartupInitializer() *startupInitializer {
	nPortCalls := uint64(0)
	return &startupInitializer{nPortCalls: &nPortCalls}
}

// getNextTigrisPort gets the next port number of Tigris to be used
// for testing in Round Robin fashion.
func (p *startupInitializer) getNextTigrisPort() uint16 {
	i := atomic.AddUint64(p.nPortCalls, 1)
	numPorts := uint64(len(ports))

	return ports[i%numPorts]
}
