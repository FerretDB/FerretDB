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

// Package pg provides PostgreSQL handler.
package pg

import (
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// Handler implements handlers.Interface on top of PostgreSQL.
type Handler struct {
	*NewOpts
}

// NewOpts represents handler configuration.
type NewOpts struct {
	PgPool        *pgdb.Pool
	L             *zap.Logger
	Metrics       *connmetrics.ConnMetrics
	StateProvider *state.Provider
}

// New returns a new handler.
func New(opts *NewOpts) (handlers.Interface, error) {
	h := &Handler{
		NewOpts: opts,
	}
	return h, nil
}

// Close implements HandlerInterface.
func (h *Handler) Close() {
	h.PgPool.Close()
}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)
