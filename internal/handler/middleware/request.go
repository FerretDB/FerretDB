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

	rw sync.RWMutex

	// we should unexport both and replace with a single wire.MsgBody field
	OpMsg   *wire.OpMsg
	OpQuery *wire.OpQuery

	doc    *wirebson.Document
	header *wire.MsgHeader
}

// RequestWire creates a new request from the given wire protocol header and body.
func RequestWire(header *wire.MsgHeader, body wire.MsgBody) *Request {
	must.NotBeZero(header)
	must.NotBeZero(body)

	req := &Request{
		header: header,
	}

	switch body := body.(type) {
	case *wire.OpMsg:
		req.OpMsg = body
	case *wire.OpQuery:
		req.OpQuery = body
	default:
		panic(fmt.Sprintf("unsupported body type %T", body))
	}

	return req
}

// RequestDoc creates a new request from the given document.
// It freezes the document.
func RequestDoc(doc *wirebson.Document) *Request {
	must.NotBeZero(doc)

	doc.Freeze()

	return &Request{
		doc: doc,
	}
}

// WireHeader returns the request header for the wire protocol.
//
// It Request was constructed with one, it is returned unmodified.
// Otherwise, a new header is created, cached, and returned.
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

	req.header = &wire.MsgHeader{
		RequestID: lastRequestID.Add(1),
	}
	var b []byte

	switch {
	case req.OpMsg != nil:
		req.header.OpCode = wire.OpCodeMsg
		b = must.NotFail(req.OpMsg.MarshalBinary())
	case req.OpQuery != nil:
		req.header.OpCode = wire.OpCodeQuery
		b = must.NotFail(req.OpQuery.MarshalBinary())
	default:
		req.header.OpCode = wire.OpCodeMsg
		req.OpMsg = must.NotFail(wire.NewOpMsg(req.doc))
		b = must.NotFail(req.OpMsg.MarshalBinary())
	}

	req.header.MessageLength = int32(wire.MsgHeaderLen + len(b))

	return req.header
}

// WireBody returns the request body for the wire protocol.
//
// If Request was constructed with one, it is returned unmodified.
// Otherwise, a new [*wire.OpMsg] is created, cached, and returned.
func (req *Request) WireBody() wire.MsgBody {
	req.rw.RLock()

	switch {
	case req.OpMsg != nil:
		req.rw.RUnlock()
		return req.OpMsg
	case req.OpQuery != nil:
		req.rw.RUnlock()
		return req.OpQuery
	}

	req.rw.RUnlock()
	req.rw.Lock()
	defer req.rw.Unlock()

	must.NotBeZero(req.doc)
	req.OpMsg = must.NotFail(wire.NewOpMsg(req.doc))

	return req.OpMsg
}

// Document returns the request document.
//
// It Request was constructed with one, it is returned unmodified.
// If request body contains a single document, it is frozen, cached, and returned.
// Otherwise, this function panics.
func (req *Request) Document() *wirebson.Document {
	req.rw.RLock()

	if req.doc != nil {
		req.rw.RUnlock()
		return req.doc
	}

	req.rw.RUnlock()
	req.rw.Lock()
	defer req.rw.Unlock()

	var doc *wirebson.Document

	switch {
	case req.OpMsg != nil:
		doc = must.NotFail(req.OpMsg.Document())
	case req.OpQuery != nil:
		doc = must.NotFail(req.OpQuery.Query())
	default:
		panic("not reached")
	}

	doc.Freeze()
	req.doc = doc

	return req.doc
}

// DocumentRaw returns the raw request document.
func (req *Request) DocumentRaw() wirebson.RawDocument {
	return must.NotFail(req.Document().Encode())
}

// lastRequestID stores last generated request ID.
var lastRequestID atomic.Int32

func init() {
	// so generated IDs are noticeably different from IDs from typical clients
	lastRequestID.Store(1_000_000_000)
}
