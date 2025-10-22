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
	"sync/atomic"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// lastRequestID stores last generated request ID.
var lastRequestID atomic.Int32

// Request represents an incoming command from the client.
//
// It may be constructed from [*wire.MsgHeader]/[wire.MsgBody] (for the wire protocol listener)
// or from [*wirebson.Document] (for data API and MCP listeners).
// The other value is extracted/generated and cached during Request creation, because we need both:
//   - raw documents carried by [wire.MsgBody] are used by both DocumentDB and proxy handlers;
//   - (non-deeply) decoded documents are used by routing, metrics, etc.
type Request struct {
	header *wire.MsgHeader
	body   wire.MsgBody
	doc    *wirebson.Document // only section 0 for OpMsg
}

// RequestWire creates a new request from the given wire protocol header and body.
// Error is returned if the body cannot be decoded.
func RequestWire(header *wire.MsgHeader, body wire.MsgBody) (*Request, error) {
	must.NotBeZero(header)
	must.NotBeZero(body)

	req := &Request{
		header: header,
		body:   body,
	}

	switch body := body.(type) {
	case *wire.OpMsg:
		if header.OpCode != wire.OpCodeMsg {
			return nil, lazyerrors.Errorf("expected OpCodeMsg, got %s", header.OpCode)
		}

		var err error
		if req.doc, err = body.Section0(); err != nil {
			return nil, lazyerrors.Error(err)
		}

	case *wire.OpQuery:
		if header.OpCode != wire.OpCodeQuery {
			return nil, lazyerrors.Errorf("expected OpCodeQuery, got %s", header.OpCode)
		}

		var err error
		if req.doc, err = body.Query(); err != nil {
			return nil, lazyerrors.Error(err)
		}

	default:
		return nil, lazyerrors.Errorf("unsupported body type %T", body)
	}

	req.doc.Freeze()

	return req, nil
}

// RequestDoc creates a new request from the given document.
// Error is returned if it cannot be decoded.
//
// If it is [*wirebson.Document], it freezes it.
func RequestDoc(doc wirebson.AnyDocument) (*Request, error) {
	must.NotBeZero(doc)

	body, err := wire.NewOpMsg(doc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	header := &wire.MsgHeader{
		MessageLength: int32(wire.MsgHeaderLen + body.Size()),
		RequestID:     lastRequestID.Add(1),
		OpCode:        wire.OpCodeMsg,
	}

	d, err := doc.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	d.Freeze()

	return &Request{
		header: header,
		body:   body,
		doc:    d,
	}, nil
}

// WireHeader returns the request header for the wire protocol.
func (req *Request) WireHeader() *wire.MsgHeader {
	return req.header
}

// WireBody returns the request body for the wire protocol.
func (req *Request) WireBody() wire.MsgBody {
	return req.body
}

// DocumentRaw returns the raw request document
// (only section 0 for OpMsg).
func (req *Request) DocumentRaw() wirebson.RawDocument {
	switch body := req.body.(type) {
	case *wire.OpMsg:
		return body.Section0Raw()
	case *wire.OpQuery:
		return body.QueryRaw()
	default:
		panic(fmt.Sprintf("unexpected body type %T", body))
	}
}

// Document returns the request document
// (only section 0 for OpMsg).
func (req *Request) Document() *wirebson.Document {
	return req.doc
}

// DocumentDeep returns the deeply decoded request document.
// Callers should use it instead of `resp.DocumentRaw().DecodeDeep()`.
func (req *Request) DocumentDeep() (*wirebson.Document, error) {
	// we might want to cache it in the future if there are many callers
	return req.DocumentRaw().DecodeDeep()
}
