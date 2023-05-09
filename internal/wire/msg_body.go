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
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
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

//go-sumtype:decl MsgBody

// ErrZeroRead is returned when zero bytes was read from connection,
// indicating that connection was closed by the client.
var ErrZeroRead = errors.New("zero bytes read")

// ReadMessage reads from reader and returns wire header and body.
//
// Error is (possibly wrapped) ErrZeroRead if zero bytes was read.
func ReadMessage(r *bufio.Reader, skipChecksum bool) (*MsgHeader, MsgBody, error) {
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

		if !skipChecksum {
			// verify message checksum if present
			flagBit := OpMsgFlags(binary.LittleEndian.Uint32(b[MsgHeaderLen : MsgHeaderLen+4]))

			if flagBit.FlagSet(OpMsgChecksumPresent) {
				msgBytes := make([]byte, header.MessageLength)

				headBytes, err := header.MarshalBinary()
				if err != nil {
					return &header, nil, lazyerrors.Error(err)
				}

				_ = append(msgBytes, headBytes...)
				msgBytes = append(msgBytes, b...)

				if err := verifyChecksum(msgBytes); err != nil {
					return &header, nil, lazyerrors.Error(err)
				}
			}
		}

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

// verifyChecksum verifies the checksum attached to an OP_MSG.
func verifyChecksum(msg []byte) error {
	table := crc32.MakeTable(crc32.Castagnoli)
	expected := binary.LittleEndian.Uint32(msg[len(msg)-crc32.Size:])
	checksum := crc32.Checksum(msg, table)

	if expected != checksum {
		return lazyerrors.New("OP_MSG checksum does not match contents")
	}

	return nil
}
