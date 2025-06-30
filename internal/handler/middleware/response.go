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
	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// Response represents outgoing result to the client.
//
// It may be constructed from [wirebson.AnyDocument] (for the DocumentDB handler),
// or from [*wire.MsgHeader] and [wire.MsgBody] (for the proxy handler).
type Response struct {
	// The order of fields is weird to make the struct smaller due to alignment.

	OpMsg   *wire.OpMsg
	OpReply *wire.OpReply
	header  *wire.MsgHeader
}

// ResponseMsg creates a new [*wire.OpMsg] response from the given document.
func ResponseMsg(doc wirebson.AnyDocument) (*Response, error) {
	msg, err := wire.NewOpMsg(doc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Response{OpMsg: msg}, nil
}

// ResponseReply creates a new [*wire.OpReply] response from the given document.
func ResponseReply(doc wirebson.AnyDocument) (*Response, error) {
	reply, err := wire.NewOpReply(doc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Response{OpReply: reply}, nil
}

// ResponseWire creates a new response from the given wire protocol header and body.
func ResponseWire(header *wire.MsgHeader, body wire.MsgBody) (*Response, error) {
	resp := &Response{
		header: header,
	}

	switch body := body.(type) {
	case *wire.OpMsg:
		if header.OpCode != wire.OpCodeMsg {
			return nil, lazyerrors.Errorf("unexpected header %s for body %T", header, body)
		}

		resp.OpMsg = body

	case *wire.OpReply:
		if header.OpCode != wire.OpCodeReply {
			return nil, lazyerrors.Errorf("unexpected header %s for body %T", header, body)
		}

		resp.OpReply = body

	default:
		return nil, lazyerrors.Errorf("unexpected response body type %T for header %s", body, header)
	}

	return resp, nil
}

// WireHeader returns the request header for the wire protocol.
func (resp *Response) WireHeader() *wire.MsgHeader {
	return resp.header
}
