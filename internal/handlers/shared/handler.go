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

package shared

import "github.com/FerretDB/FerretDB/internal/pg"

// Handler data struct.
type Handler struct {
	pgPool   *pg.Pool
	peerAddr string
}

// NewHandler returns a pointer to a new Handler, populated with the pgPool and peerAddr.
func NewHandler(pgPool *pg.Pool, peerAddr string) *Handler {
	return &Handler{
		pgPool:   pgPool,
		peerAddr: peerAddr,
	}
}
