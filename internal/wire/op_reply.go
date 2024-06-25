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

package wire

import (
	"encoding/binary"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// OpReply is a deprecated response message type.
//
// Only up to one returned document is supported.
type OpReply struct {
	// The order of fields is weird to make the struct smaller due to alignment.
	// The wire order is: flags, cursor ID, starting from, documents.

	document     bson.RawDocument
	CursorID     int64
	Flags        OpReplyFlags
	StartingFrom int32
}

// NewOpReply creates a new OpReply message.
func NewOpReply(doc bson.AnyDocument) (*OpReply, error) {
	raw, err := doc.Encode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &OpReply{document: raw}, nil
}

func (reply *OpReply) msgbody() {}

// check implements [MsgBody] interface.
func (reply *OpReply) check() error {
	if d := reply.document; d != nil {
		if _, err := d.DecodeDeep(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// UnmarshalBinaryNocopy implements [MsgBody] interface.
func (reply *OpReply) UnmarshalBinaryNocopy(b []byte) error {
	if len(b) < 20 {
		return lazyerrors.Errorf("len=%d", len(b))
	}

	reply.Flags = OpReplyFlags(binary.LittleEndian.Uint32(b[0:4]))
	reply.CursorID = int64(binary.LittleEndian.Uint64(b[4:12]))
	reply.StartingFrom = int32(binary.LittleEndian.Uint32(b[12:16]))
	numberReturned := int32(binary.LittleEndian.Uint32(b[16:20]))
	reply.document = b[20:]

	if numberReturned < 0 || numberReturned > 1 {
		return lazyerrors.Errorf("numberReturned=%d", numberReturned)
	}

	if len(reply.document) == 0 {
		reply.document = nil
	}

	if (numberReturned == 0) != (reply.document == nil) {
		return lazyerrors.Errorf("numberReturned=%d, document=%v", numberReturned, reply.document)
	}

	if debugbuild.Enabled {
		if err := reply.check(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// MarshalBinary implements [MsgBody] interface.
func (reply *OpReply) MarshalBinary() ([]byte, error) {
	if debugbuild.Enabled {
		if err := reply.check(); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	b := make([]byte, 20+len(reply.document))

	binary.LittleEndian.PutUint32(b[0:4], uint32(reply.Flags))
	binary.LittleEndian.PutUint64(b[4:12], uint64(reply.CursorID))
	binary.LittleEndian.PutUint32(b[12:16], uint32(reply.StartingFrom))

	if reply.document == nil {
		binary.LittleEndian.PutUint32(b[16:20], uint32(0))
	} else {
		binary.LittleEndian.PutUint32(b[16:20], uint32(1))
		copy(b[20:], reply.document)
	}

	return b, nil
}

// Document returns reply document.
func (reply *OpReply) Document() (*types.Document, error) {
	if reply.document == nil {
		return nil, nil
	}

	return reply.document.Convert()
}

// SetDocument sets reply document.
func (reply *OpReply) SetDocument(doc *types.Document) {
	d := must.NotFail(bson.ConvertDocument(doc))
	reply.document = must.NotFail(d.Encode())
}

// logMessage returns a string representation for logging.
func (reply *OpReply) logMessage(logFunc func(v any) string) string {
	if reply == nil {
		return "<nil>"
	}

	m := must.NotFail(bson.NewDocument(
		"ResponseFlags", reply.Flags.String(),
		"CursorID", reply.CursorID,
		"StartingFrom", reply.StartingFrom,
	))

	if reply.document == nil {
		must.NoError(m.Add("NumberReturned", int32(0)))
	} else {
		must.NoError(m.Add("NumberReturned", int32(1)))

		doc, err := reply.document.DecodeDeep()
		if err == nil {
			must.NoError(m.Add("Document", doc))
		} else {
			must.NoError(m.Add("DocumentError", err.Error()))
		}
	}

	return logFunc(m)
}

// String returns a string representation for logging.
func (reply *OpReply) String() string {
	return reply.logMessage(bson.LogMessage)
}

// StringBlock returns an indented string representation for logging.
func (reply *OpReply) StringBlock() string {
	return reply.logMessage(bson.LogMessageBlock)
}

// StringFlow returns an unindented string representation for logging.
func (reply *OpReply) StringFlow() string {
	return reply.logMessage(bson.LogMessageFlow)
}

// check interfaces
var (
	_ MsgBody = (*OpReply)(nil)
)
