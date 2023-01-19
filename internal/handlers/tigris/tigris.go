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

// Package tigris provides Tigris handler.
package tigris

import (
	"context"

	"github.com/tigrisdata/tigris-client-go/config"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/connmetrics"
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// NewOpts represents handler configuration.
type NewOpts struct {
	ClientID      string
	ClientSecret  string
	Token         string
	URL           string
	L             *zap.Logger
	Metrics       *connmetrics.ConnMetrics
	StateProvider *state.Provider
}

// Handler implements handlers.Interface on top of Tigris.
type Handler struct {
	*NewOpts

	// TODO replace with map
	// https://github.com/FerretDB/FerretDB/issues/1789
	db *tigrisdb.TigrisDB
}

// New returns a new handler.
func New(opts *NewOpts) (handlers.Interface, error) {
	if opts.URL == "" {
		return nil, lazyerrors.New("Tigris URL is not provided")
	}

	cfg := &config.Driver{
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		Token:        opts.Token,
		URL:          opts.URL,
	}
	db, err := tigrisdb.New(context.TODO(), cfg, opts.L, true)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h := &Handler{
		NewOpts: opts,
		db:      db,
	}

	return h, nil
}

// Close implements handlers.Interface.
func (h *Handler) Close() {
	h.db.Driver.Close()
}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)
