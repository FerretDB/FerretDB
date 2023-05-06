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
	"bufio"
	"encoding"
	"errors"
	"fmt"
	"io"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// MsgBody is a wire protocol message body.
type MsgBody interface {
	readFrom(*bufio.Reader) error
	encoding.BinaryUnmarshaler
	encoding.BinaryMarshaler
	fmt.Stringer

	msgbody() // seal for go-sumtype
}

// crc32c checksum byte size
const kCrc32Size = 4

//go-sumtype:decl MsgBody

// ErrZeroRead is returned when zero bytes was read from connection,
// indicating that connection was closed by the client.
var ErrZeroRead = errors.New("zero bytes read")

// ReadMessage reads from reader and returns wire header and body.
//
// Error is (possibly wrapped) ErrZeroRead if zero bytes was read.
func ReadMessage(r *bufio.Reader) (*MsgHeader, MsgBody, error) {

	var header MsgHeader
	if err := header.readFrom(r); err != nil {
		return nil, nil, lazyerrors.Error(err)
	}

	b := make([]byte, header.MessageLength-MsgHeaderLen)
	if n, err := io.ReadFull(r, b); err != nil {
		return nil, nil, lazyerrors.Errorf("expected %d, read %d: %w", len(b), n, err)
	}

	switch header.OpCode {
	case OpCodeReply: // not sent by clients, but we should be able to read replies from a proxy
		var reply OpReply
		if err := reply.UnmarshalBinary(b); err != nil {
			return nil, nil, lazyerrors.Error(err)
		}

		return &header, &reply, nil

	case OpCodeMsg:
		var msg OpMsg
		if err := msg.UnmarshalBinary(b); err != nil {
			return &header, nil, lazyerrors.Error(err)
		}

		return &header, &msg, nil

	case OpCodeQuery:
		var query OpQuery
		if err := query.UnmarshalBinary(b); err != nil {
			return nil, nil, lazyerrors.Error(err)
		}

		return &header, &query, nil

	case OpCodeUpdate:
		fallthrough
	case OpCodeInsert:
		fallthrough
	case OpCodeGetByOID:
		fallthrough
	case OpCodeGetMore:
		fallthrough
	case OpCodeDelete:
		fallthrough
	case OpCodeKillCursors:
		fallthrough
	case OpCodeCompressed:
		return nil, nil, lazyerrors.Errorf("unhandled opcode %s", header.OpCode)

	default:
		return nil, nil, lazyerrors.Errorf("unexpected opcode %s", header.OpCode)
	}
}

// WriteMessage validates msg and headers and writes them to the writer.
func WriteMessage(w *bufio.Writer, header *MsgHeader, msg MsgBody) error {
	b, err := msg.MarshalBinary()
	if err != nil {
		return lazyerrors.Error(err)
	}

	if expected := len(b) + MsgHeaderLen; int32(expected) != header.MessageLength {
		panic(fmt.Sprintf(
			"expected length %d (marshaled body size) + %d (fixed marshaled header size) = %d, got %d",
			len(b), MsgHeaderLen, expected, header.MessageLength,
		))
	}

	if err := header.writeTo(w); err != nil {
		return lazyerrors.Error(err)
	}

	if _, err := w.Write(b); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
