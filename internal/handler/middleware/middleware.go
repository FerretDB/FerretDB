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

// Package middleware provides wrappers for command handlers.
package middleware

import (
	"context"
	"sync/atomic"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Handler is a common interface for command handlers.
type Handler interface {
	// Handle processes a single request.
	//
	// The passed context is canceled when the client disconnects.
	//
	// Response is a normal response or an error.
	// TODO https://github.com/FerretDB/FerretDB/issues/4965
	Handle(ctx context.Context, req *Request) (resp *Response, err error)
}

// lastRequestID stores last generated request ID.
var lastRequestID atomic.Int32

type Middleware struct {
	opts *NewOpts
}

type NewOpts struct {
	Mode     Mode
	Handlers []Handler
}

func New(opts *NewOpts) *Middleware {
	must.NotBeZero(opts)

	return &Middleware{
		opts: opts,
	}
}

func (m *Middleware) Handle(ctx context.Context, req *Request) (resp *Response, err error) {
	panic("TODO")
}

// check interfaces
var (
	_ Handler = (*Middleware)(nil)
)
