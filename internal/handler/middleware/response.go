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

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Response is a normal or error response produced by the handler.
//
// It may be constructed from [*wire.MsgHeader]/[wire.MsgBody] (for the proxy handler),
// from [wirebson.AnyDocument] (for the DocumentDB handler),
// or from [*mongoerrors.Error] (for both).
type Response struct {
	header     *wire.MsgHeader
	body       wire.MsgBody
	doc        *wirebson.Document // only section 0 for OpMsg
	mongoError *mongoerrors.Error // may be nil even for error response (for proxy handler, for example)
}

// ResponseWire creates a new response from the given wire protocol header and body.
// Error is returned if the body cannot be decoded.
func ResponseWire(header *wire.MsgHeader, body wire.MsgBody) (*Response, error) {
	must.NotBeZero(header)
	must.NotBeZero(body)

	resp := &Response{
		header: header,
		body:   body,
	}

	switch body := body.(type) {
	case *wire.OpMsg:
		if header.OpCode != wire.OpCodeMsg {
			return nil, lazyerrors.Errorf("expected OpCodeMsg, got %s", header.OpCode)
		}

		var err error
		if resp.doc, err = body.Section0(); err != nil {
			return nil, lazyerrors.Error(err)
		}

	case *wire.OpReply:
		if header.OpCode != wire.OpCodeReply {
			return nil, lazyerrors.Errorf("expected OpCodeReply, got %s", header.OpCode)
		}

		var err error
		if resp.doc, err = body.Document(); err != nil {
			return nil, lazyerrors.Error(err)
		}

	default:
		return nil, lazyerrors.Errorf("unsupported body type %T", body)
	}

	resp.doc.Freeze()

	return resp, nil
}

// ResponseDoc creates a new response from the given document.
// Error is returned if it cannot be decoded.
//
// If it is [*wirebson.Document], it freezes it.
func ResponseDoc(req *Request, doc wirebson.AnyDocument) (*Response, error) {
	must.NotBeZero(req)
	must.NotBeZero(doc)

	resp := &Response{
		header: &wire.MsgHeader{
			RequestID:  lastRequestID.Add(1),
			ResponseTo: req.WireHeader().RequestID,
		},
	}

	switch req.WireHeader().OpCode {
	case wire.OpCodeMsg:
		resp.header.OpCode = wire.OpCodeMsg

		var err error
		if resp.body, err = wire.NewOpMsg(doc); err != nil {
			return nil, lazyerrors.Error(err)
		}

	case wire.OpCodeQuery:
		resp.header.OpCode = wire.OpCodeReply

		var err error
		if resp.body, err = wire.NewOpReply(doc); err != nil {
			return nil, lazyerrors.Error(err)
		}

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
		return nil, lazyerrors.Errorf("unexpected request header %s", req.WireHeader())
	}

	resp.header.MessageLength = int32(wire.MsgHeaderLen + resp.body.Size())

	var err error
	if resp.doc, err = doc.Decode(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	resp.doc.Freeze()

	return resp, nil
}

// ResponseErr creates a new error response from the given [*mongoerrors.Error].
func ResponseErr(req *Request, err *mongoerrors.Error) *Response {
	must.NotBeZero(req)
	must.NotBeZero(err)

	doc := wirebson.MustDocument(
		"ok", float64(0),
		"errmsg", err.Message,
		"code", err.Code,
		"codeName", err.Name,
	)

	resp := must.NotFail(ResponseDoc(req, doc))
	resp.mongoError = err

	return resp
}

// WireHeader returns the response header for the wire protocol.
func (resp *Response) WireHeader() *wire.MsgHeader {
	return resp.header
}

// WireBody returns the response body for the wire protocol.
func (resp *Response) WireBody() wire.MsgBody {
	return resp.body
}

// DocumentRaw returns the raw response document
// (only section 0 for OpMsg).
func (resp *Response) DocumentRaw() wirebson.RawDocument {
	switch body := resp.body.(type) {
	case *wire.OpMsg:
		return body.Section0Raw()
	case *wire.OpReply:
		return body.DocumentRaw()
	default:
		panic(fmt.Sprintf("unexpected body type %T", body))
	}
}

// Document returns the response document
// (only section 0 for OpMsg).
func (resp *Response) Document() *wirebson.Document {
	return resp.doc
}

// DocumentDeep returns the deeply decoded response document
// (only section 0 for OpMsg).
// Callers should use it instead of `resp.DocumentRaw().DecodeDeep()`.
func (resp *Response) DocumentDeep() (*wirebson.Document, error) {
	// we might want to cache it in the future if there are many callers
	return resp.DocumentRaw().DecodeDeep()
}

// OK returns true if the response document contains the "ok" field with numeric value 1
// (and, temporarily, boolean value true).
func (resp *Response) OK() bool {
	switch v := resp.doc.Get("ok").(type) {
	case float64:
		return v == float64(1)
	case bool:
		// TODO https://github.com/documentdb/documentdb/issues/49
		return v
	case int32:
		return v == int32(1)
	case int64:
		return v == int64(1)
	default:
		return false
	}
}

// ErrorCode returns the error code of the response, or [mongoerrors.ErrUnset] (zero).
func (resp *Response) ErrorCode() mongoerrors.Code {
	if resp.mongoError != nil {
		return mongoerrors.Code(resp.mongoError.Code)
	}

	if resp.OK() {
		return mongoerrors.ErrUnset
	}

	switch v := resp.doc.Get("code").(type) {
	case int32:
		return mongoerrors.Code(v)
	case int64:
		return mongoerrors.Code(v)
	default:
		return mongoerrors.ErrUnset
	}
}

// ErrorName returns the error name of the response, or an empty string.
func (resp *Response) ErrorName() string {
	if resp.mongoError != nil {
		return resp.mongoError.Name
	}

	if resp.OK() {
		return ""
	}

	if v := resp.doc.Get("codeName"); v != nil {
		n, _ := v.(string)
		return n
	}

	// legacy replies do not have "codeName" field
	if c := resp.ErrorCode(); c != mongoerrors.ErrUnset {
		return c.String()
	}

	return ""
}

// MongoError returns [*mongoerrors.Error], if the response was created with [ResponseErr].
func (resp *Response) MongoError() *mongoerrors.Error {
	return resp.mongoError
}
