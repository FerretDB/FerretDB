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
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// MsgRequest represents incoming request from the client.
// It may come from the wire protocol connection or from the Data API server.
type MsgRequest struct {
	*wire.OpMsg
	RequestID int32
	Command   string
}

// MsgResponse represent outgoing response to the client.
type MsgResponse struct {
	*wire.OpMsg
	*mongoerrors.Error
}

// QueryRequest is a deprecated request message type.
// It is still used by commands including `hello` and `isMaster`.
type QueryRequest struct {
	*wire.OpQuery
	RequestID int32
}

// ReplyResponse is a deprecated response message type used for the response to [QueryRequest].
type ReplyResponse struct {
	*wire.OpReply
	*mongoerrors.Error
}

// Middleware represents functions for handling incoming requests.
type Middleware interface {
	HandleOpMsg(next MsgHandlerFunc) MsgHandlerFunc
	HandleOpReply(next QueryHandlerFunc) QueryHandlerFunc
}

// Response constructs a [*MsgResponse] from a single document.
func Response(doc wirebson.AnyDocument) (*MsgResponse, error) {
	msg, err := wire.NewOpMsg(doc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &MsgResponse{OpMsg: msg}, nil
}

// Reply constructs a [*ReplyResponse] from a single document.
func Reply(doc wirebson.AnyDocument) (*ReplyResponse, error) {
	reply, err := wire.NewOpReply(doc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &ReplyResponse{OpReply: reply}, nil
}

// MsgHandlerFunc represents a function/method that processes a single OP_MSG command.
//
// The passed context is canceled when the client disconnects.
type MsgHandlerFunc func(ctx context.Context, req *MsgRequest) (resp *MsgResponse, err error)

// QueryHandlerFunc represents a function/method that processes a single OP_QUERY command.
//
// The passed context is canceled when the client disconnects.
type QueryHandlerFunc func(ctx context.Context, req *QueryRequest) (resp *ReplyResponse, err error)
