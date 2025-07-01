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
	"sync"
	"sync/atomic"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Request represents incoming command from the client.
//
// It may be constructed from [*wire.MsgHeader] and [wire.MsgBody] (for the wire protocol listener)
// or from [*wirebson.Document] (for data API and MCP listeners).
type Request struct {
	// The order of fields is weird to make the struct smaller due to alignment.
	// All fields are protected by rw.

	rw     sync.RWMutex
	header *wire.MsgHeader
	body   wire.MsgBody
	doc    *wirebson.Document
	raw    wirebson.RawDocument
}

// RequestWire creates a new request from the given wire protocol header and body.
func RequestWire(header *wire.MsgHeader, body wire.MsgBody) *Request {
	must.NotBeZero(header)
	must.NotBeZero(body)

	req := &Request{
		header: header,
	}

	switch body := body.(type) {
	case *wire.OpMsg, *wire.OpQuery:
		req.body = body
	default:
		panic(fmt.Sprintf("unsupported body type %T", body))
	}

	return req
}

// RequestDoc creates a new request from the given document.
// If it is [*wirebson.Document], it freezes it.
func RequestDoc(doc wirebson.AnyDocument) *Request {
	must.NotBeZero(doc)

	switch d := doc.(type) {
	case *wirebson.Document:
		d.Freeze()

		return &Request{
			doc: d,
		}
	case wirebson.RawDocument:
		return &Request{
			raw: d,
		}
	default:
		panic("not reached")
	}
}

// setHeaderBody ensures that header and body are set.
func (req *Request) setHeaderBody() {
	if req.header != nil && req.body != nil {
		return
	}

	// wire.OpMsg always uses the raw document under the hood,
	// so store it to avoid encoding it again
	if req.raw == nil {
		must.NotBeZero(req.doc)
		req.raw = must.NotFail(req.doc.Encode())
	}

	req.body = must.NotFail(wire.NewOpMsg(req.raw))

	b := must.NotFail(req.body.MarshalBinary())

	req.header = &wire.MsgHeader{
		MessageLength: int32(wire.MsgHeaderLen + len(b)),
		RequestID:     lastRequestID.Add(1),
		OpCode:        wire.OpCodeMsg,
	}
}

// WireHeader returns the request header for the wire protocol.
//
// It Request was constructed with one, it is returned unmodified.
// Otherwise, a new header is created, cached, and returned.
// It panics if request body can't be marshaled.
func (req *Request) WireHeader() *wire.MsgHeader {
	req.rw.RLock()

	if req.header != nil {
		req.rw.RUnlock()
		return req.header
	}

	req.rw.RUnlock()
	req.rw.Lock()
	defer req.rw.Unlock()

	// a concurrent call might have set it already; check again
	if req.header != nil {
		return req.header
	}

	req.setHeaderBody()

	return req.header
}

// WireBody returns the request body for the wire protocol.
//
// If Request was constructed with one, it is returned unmodified.
// Otherwise, a new [*wire.OpMsg] is created, cached, and returned.
func (req *Request) WireBody() wire.MsgBody {
	req.rw.RLock()

	if req.body != nil {
		req.rw.RUnlock()
		return req.body
	}

	req.rw.RUnlock()
	req.rw.Lock()
	defer req.rw.Unlock()

	// a concurrent call might have set it already; check again
	if req.body != nil {
		return req.body
	}

	req.setHeaderBody()

	return req.body
}

// setRaw ensures that raw is set.
func (req *Request) setRaw() {
	if req.raw != nil {
		return
	}

	switch body := req.body.(type) {
	case *wire.OpMsg:
		req.raw = body.Section0Raw()
	case *wire.OpQuery:
		req.raw = body.QueryRaw()
	default:
		must.NotBeZero(req.doc)
		req.raw = must.NotFail(req.doc.Encode())
	}
}

// Document returns the request document.
//
// It Request was constructed with one, it is returned unmodified.
// Otherwise, a new [*wirebson.Document] is created from the request body (section 0 only for [wire.OpMsg]),
// frozen, cached, and returned.
func (req *Request) Document() *wirebson.Document {
	req.rw.RLock()

	if req.doc != nil {
		req.rw.RUnlock()
		return req.doc
	}

	req.rw.RUnlock()
	req.rw.Lock()
	defer req.rw.Unlock()

	// a concurrent call might have set it already; check again
	if req.doc != nil {
		return req.doc
	}

	if req.raw == nil {
		req.setRaw()
	}

	req.doc = must.NotFail(req.raw.Decode())
	req.doc.Freeze()
	return req.doc
}

// DocumentRaw returns the raw request document.
//
// It Request was constructed with one, it is returned unmodified.
// Otherwise, a new [wirebson.RawDocument] is created from the request body (section 0 only for [wire.OpMsg]),
// cached, and returned.
func (req *Request) DocumentRaw() wirebson.RawDocument {
	req.rw.RLock()

	if req.raw != nil {
		req.rw.RUnlock()
		return req.raw
	}

	req.rw.RUnlock()
	req.rw.Lock()
	defer req.rw.Unlock()

	// a concurrent call might have set it already; check again
	if req.raw != nil {
		return req.raw
	}

	req.setRaw()

	return req.raw
}

// lastRequestID stores last generated request ID.
var lastRequestID atomic.Int32

func init() {
	// so generated IDs are noticeably different from IDs from typical clients
	lastRequestID.Store(1_000_000_000)
}
