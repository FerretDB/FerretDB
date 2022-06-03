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
	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
)

// errNotImplemented is returned by stubs.
var errNotImplemented = common.NewErrorMsg(common.ErrNotImplemented, "This command is not implemented for Tigris yet")

// Handler implements handlers.Interface on top of Tigris.
type Handler struct{}

// New returns a new handler.
func New() handlers.Interface {
	return new(Handler)
}

// Close implements handlers.Interface.
func (h *Handler) Close() {}

// check interfaces
var (
	_ handlers.Interface = (*Handler)(nil)
)
