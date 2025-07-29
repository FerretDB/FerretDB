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

package wiring

import (
	"context"
	"log/slog"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

type SetupOpts struct {
	StateProvider *state.Provider
	PostgreSQLURL string
	Logger        *slog.Logger
}

type SetupResult struct{}

func Setup(ctx context.Context, opts *SetupOpts) *SetupResult {
	must.NotBeZero(opts)

	p, err := documentdb.NewPool(opts.PostgreSQLURL, logging.WithName(opts.Logger, "pool"), opts.StateProvider)
	if err != nil {
		opts.Logger.LogAttrs(ctx, logging.LevelDPanic, "Failed to construct connection pool", logging.Error(err))
		return nil
	}

	_ = p

	return &SetupResult{}
}
