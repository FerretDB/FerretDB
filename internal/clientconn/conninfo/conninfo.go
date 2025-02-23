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

// Package conninfo provides access to connection-specific information.
package conninfo

import (
	"net/netip"
	"sync"

	"github.com/FerretDB/FerretDB/v2/internal/util/scram"
)

// ConnInfo represents client connection information.
type ConnInfo struct {
	// the order of fields is weird to make the struct smaller due to alignment

	conv         *scram.Conv    // protected by rw
	Peer         netip.AddrPort // invalid for Unix domain sockets
	rw           sync.RWMutex   // rw
	metadataRecv bool           // protected by rw
	steps        int            // protected by rw
}

// New creates a new ConnInfo.
func New() *ConnInfo {
	return new(ConnInfo)
}

// Conv returns SCRAM conversation.
func (ci *ConnInfo) Conv() *scram.Conv {
	ci.rw.RLock()
	defer ci.rw.RUnlock()

	return ci.conv
}

// SetConv sets SCRAM conversation.
// It returns true if existing conversation was replaced.
func (ci *ConnInfo) SetConv(conv *scram.Conv) bool {
	ci.rw.Lock()
	defer ci.rw.Unlock()

	was := ci.conv != nil
	ci.conv = conv

	return was
}

// MetadataRecv returns whatever client metadata was received already.
func (ci *ConnInfo) MetadataRecv() bool {
	ci.rw.RLock()
	defer ci.rw.RUnlock()

	return ci.metadataRecv
}

// SetMetadataRecv marks client metadata as received.
func (ci *ConnInfo) SetMetadataRecv() {
	ci.rw.Lock()
	defer ci.rw.Unlock()

	ci.metadataRecv = true
}

// DecrementSteps decreases the steps counter and returns the number of steps left
// to complete the handshake.
// The final step returns `0`, a completed handshake returns a negative value.
func (ci *ConnInfo) DecrementSteps() int {
	ci.rw.Lock()
	defer ci.rw.Unlock()

	ci.steps--

	return ci.steps
}

// SetSteps sets the number of round trips left to complete the handshake.
func (ci *ConnInfo) SetSteps(steps int) {
	ci.rw.Lock()
	defer ci.rw.Unlock()

	ci.steps = steps
}
