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
	"fmt"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Response represents outgoing result to the client.
// It maybe a normal response or an error.
// TODO https://github.com/FerretDB/FerretDB/issues/4965
//
// It may be constructed from [wirebson.AnyDocument] (for the DocumentDB handler),
// or from [*wire.MsgHeader] and [wire.MsgBody] (for the proxy handler).
type Response struct {
	header *wire.MsgHeader
	body   wire.MsgBody

	// Remove, replace with body.
	// TODO https://github.com/FerretDB/FerretDB/issues/4965
	OpMsg *wire.OpMsg
}

// ResponseDoc creates a new response from the given document.
func ResponseDoc(req *Request, doc wirebson.AnyDocument) (*Response, error) {
	reqHeader := req.WireHeader()
	res := &Response{
		header: &wire.MsgHeader{
			RequestID:  lastRequestID.Add(1),
			ResponseTo: reqHeader.ResponseTo,
		},
	}

	var err error

	switch reqHeader.OpCode {
	case wire.OpCodeMsg:
		res.header.OpCode = wire.OpCodeMsg
		res.OpMsg, err = wire.NewOpMsg(doc)
		res.body = res.OpMsg
	case wire.OpCodeQuery:
		res.header.OpCode = wire.OpCodeReply
		res.body, err = wire.NewOpReply(doc)
	case wire.OpCodeReply:
		fallthrough
	case wire.OpCodeUpdate:
		fallthrough
	case wire.OpCodeInsert:
		fallthrough
	case wire.OpCodeGetByOID:
		fallthrough
	case wire.OpCodeGetMore:
		fallthrough
	case wire.OpCodeDelete:
		fallthrough
	case wire.OpCodeKillCursors:
		fallthrough
	case wire.OpCodeCompressed:
		fallthrough
	default:
		return nil, lazyerrors.Errorf("unexpected request header %s", reqHeader)
	}
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res.header.MessageLength = int32(wire.MsgHeaderLen + res.body.Size())

	return res, nil
}

// ResponseWire creates a new response from the given wire protocol header and body.
func ResponseWire(header *wire.MsgHeader, body wire.MsgBody) *Response {
	must.NotBeZero(header)
	must.NotBeZero(body)

	resp := &Response{
		header: header,
	}

	switch body := body.(type) {
	case *wire.OpMsg:
		must.BeTrue(header.OpCode == wire.OpCodeMsg)
		resp.OpMsg = body
		resp.body = body
	case *wire.OpReply:
		must.BeTrue(header.OpCode == wire.OpCodeReply)
		resp.body = body
	default:
		panic(fmt.Sprintf("unsupported body type %T", body))
	}

	return resp
}

// WireHeader returns the request header for the wire protocol.
func (resp *Response) WireHeader() *wire.MsgHeader {
	return resp.header
}

// WireBody returns the request body for the wire protocol.
func (resp *Response) WireBody() wire.MsgBody {
	return resp.body
}
