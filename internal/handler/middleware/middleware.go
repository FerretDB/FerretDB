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

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// Request represents incoming request from the client.
type Request struct {
	*wire.OpMsg
	OpQuery *wire.OpQuery
}

// Response represent outgoing response to the client.
type Response struct {
	OpMsg   *wire.OpMsg
	OpReply *wire.OpReply
}

// MakeResponse constructs an OP_MSG [*Response] from a single document.
func MakeResponse(doc wirebson.AnyDocument) (*Response, error) {
	msg, err := wire.NewOpMsg(doc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Response{OpMsg: msg}, nil
}

// MakeReply constructs an OP_QUERY [*Response] from a single document.
func MakeReply(doc wirebson.AnyDocument) (*Response, error) {
	reply, err := wire.NewOpReply(doc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Response{OpReply: reply}, nil
}

// HandleFunc represents a function/method that processes a single request.
//
// The passed context is canceled when the client disconnects.
type HandleFunc func(ctx context.Context, req *Request) (resp *Response, err error)
