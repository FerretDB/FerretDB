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

// Package pool provides access to PostgreSQL connections.
//
// It should be used only by the metadata package.
package pool

import (
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/resource"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Pool provides access to PostgreSQL connections.
type Pool struct {
	u  string
	l  *zap.Logger
	sp *state.Provider

	token *resource.Token

	rw    sync.RWMutex
	pools map[string]*pgxpool.Pool
}

func New(u string, l *zap.Logger, sp *state.Provider) *Pool {
	p := &Pool{
		u:     u,
		l:     l,
		sp:    sp,
		token: resource.NewToken(),
		pools: make(map[string]*pgxpool.Pool),
	}

	resource.Track(p, p.token)

	return p
}

func (p *Pool) Close() {
	p.rw.Lock()
	defer p.rw.Unlock()

	for _, pool := range p.pools {
		pool.Close()
	}

	p.pools = nil

	resource.Untrack(p, p.token)
}
