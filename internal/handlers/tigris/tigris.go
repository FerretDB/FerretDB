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
	"time"

	"github.com/tigrisdata/tigris-client-go/config"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// notImplemented returns error for stub command handlers.
func notImplemented(command string) error {
	return common.NewErrorMsg(common.ErrNotImplemented, "I'm a stub, not a real handler for "+command)
}

// NewOpts represents handler configuration.
type NewOpts struct {
	TigrisURL string
	L         *zap.Logger
}

// Handler implements handlers.Interface on top of Tigris.
type Handler struct {
	*NewOpts
	db        *tigrisdb.TigrisDB
	startTime time.Time
}

// New returns a new handler.
func New(opts *NewOpts) (handlers.Interface, error) {
	cfg := &config.Driver{
		URL: opts.TigrisURL,
	}
	db, err := tigrisdb.New(cfg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h := &Handler{
		NewOpts:   opts,
		db:        db,
		startTime: time.Now(),
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
