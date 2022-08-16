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
	"time"

	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
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
	driver    driver.Driver
	startTime time.Time
}

// New returns a new handler.
func New(opts *NewOpts) (handlers.Interface, error) {
	cfg := &config.Driver{
		URL: opts.TigrisURL,
	}
	driver, err := driver.NewDriver(context.TODO(), cfg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h := &Handler{
		NewOpts:   opts,
		driver:    driver,
		startTime: time.Now(),
	}
	return h, nil
}

// Close implements handlers.Interface.
func (h *Handler) Close() {
	h.driver.Close()
}

// isNotFound returns true if the error is a "not found" error.
// This function is implemented to keep nolint in a single place.
func isNotFound(err *driver.Error) bool {
	if err == nil {
		return false
	}

	//nolint:nosnakecase // Tigris named their const that way
	if err.Code == api.Code_NOT_FOUND {
		return true
	}
	return false
}

// isInvalidArgument returns true if the error is a "invalid argument" error.
// This function is implemented to keep nolint in a single place.
func isInvalidArgument(err *driver.Error) bool {
	if err == nil {
		return false
	}

	//nolint:nosnakecase // Tigris named their const that way
	if err.Code == api.Code_INVALID_ARGUMENT {
		return true
	}
	return false
}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)
