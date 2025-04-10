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
	"sync"
	"sync/atomic"

	"github.com/FerretDB/wire"
)

// Request represents incoming command from the client.
type Request struct {
	// the order of fields is weird to make the struct smaller due to alignment

	OpMsg   *wire.OpMsg
	OpQuery *wire.OpQuery

	header *wire.MsgHeader // protected by rw
	rw     sync.RWMutex
}

// RequestWire creates a new request from the given wire protocol header and body.
func RequestWire(header *wire.MsgHeader, body wire.MsgBody) *Request {
	req := &Request{
		header: header,
	}

	switch body := body.(type) {
	case *wire.OpMsg:
		req.OpMsg = body
	case *wire.OpQuery:
		req.OpQuery = body
	default:
		panic("unsupported body type")
	}

	return req
}

// WireHeader returns the request header for the wire protocol.
//
// It Request was constructed with one, it is returned unmodified.
// Otherwise, a new header is created, cached, and returned.
func (req *Request) WireHeader() *wire.MsgHeader {
	req.rw.RLock()

	if h := req.header; h != nil {
		defer req.rw.RUnlock()
		return h
	}

	req.rw.RUnlock()
	req.rw.Lock()
	defer req.rw.Unlock()

	// a concurrent call might have set it already; check again
	if h := req.header; h != nil {
		return h
	}

	req.header = &wire.MsgHeader{
		RequestID: lastRequestID.Add(1),
	}
	var b []byte

	switch {
	case req.OpMsg != nil:
		req.header.OpCode = wire.OpCodeMsg
		b, _ = req.OpMsg.MarshalBinary()
	case req.OpQuery != nil:
		req.header.OpCode = wire.OpCodeQuery
		b, _ = req.OpQuery.MarshalBinary()
	default:
		panic("unsupported body type")
	}

	req.header.MessageLength = int32(wire.MsgHeaderLen + len(b))

	return req.header
}

// WireBody returns the request body for the wire protocol.
func (req *Request) WireBody() wire.MsgBody {
	switch {
	case req.OpMsg != nil:
		return req.OpMsg
	case req.OpQuery != nil:
		return req.OpQuery
	default:
		panic("empty body")
	}
}

// lastRequestID stores last generated request ID.
var lastRequestID atomic.Int32

func init() {
	// so generated IDs are noticeably different from IDs from typical clients
	lastRequestID.Store(1_000_000_000)
}
