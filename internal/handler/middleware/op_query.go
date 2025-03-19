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

package middleware

import (
	"context"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// CmdQuery represents an OP_QUERY request.
type CmdQuery struct {
	*wire.OpQuery
}

// CmdReply represents an OP_REPLY response.
type CmdReply struct {
	*wire.OpReply
}

// NewReply creates a new OP_REPLY response.
func NewReply(doc wirebson.AnyDocument) (*CmdReply, error) {
	reply, err := wire.NewOpReply(doc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &CmdReply{OpReply: reply}, nil
}

// QueryHandlerFunc represents a function/method that processes a single OP_QUERY command.
//
// The passed context is canceled when the client disconnects.
type QueryHandlerFunc func(ctx context.Context, req *CmdQuery) (resp *CmdReply, err error)
