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
func ReadMessage(r *bufio.Reader) (*MsgHeader, MsgBody, error) {

	reader := io.TeeReader(r)

	// detach checksum
	//

	if err := verifyChecksum(r); err != nil {
		return nil, nil, lazyerrors.Error(err)
	}

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

// verifyChecksum verifies the checksum of the message it is attached
func verifyChecksum(r *bufio.Reader) error {

	n := MsgHeaderLen + crc32.Size
	msgHeader, err := r.Peek(n)
	if err != nil {
		if err == io.EOF {
			return ErrZeroRead
		}
	}
	msgLen := int(binary.LittleEndian.Uint32(msgHeader[0:4]))

	if msgLen < MsgHeaderLen || msgLen > MaxMsgLen {
		return lazyerrors.Errorf("invalid message length %d", msgLen)
	}

	b, err := r.Peek(msgLen)
	if err != nil {
		return lazyerrors.Error(err)
	}

	flagbits := OpMsgFlags(binary.LittleEndian.Uint32(msgHeader[MsgHeaderLen:n]))
	if flagbits.FlagSet(OpMsgChecksumPresent) {
		// remove checksum from the message
		actualMsg, checksum := detachChecksum(b)

		if checksum != calculateChecksum(actualMsg) {
			return lazyerrors.New("OP_MSG checksum does not match contents")
		}
	}

	return nil
}

// attachChecksum appends checksum to a message
func attachChecksum(data []byte) []byte {
	var checksum []byte
	binary.LittleEndian.PutUint32(checksum, calculateChecksum(data))
	return checksum
}

// detachChecksum removes the checksum bytes from a message
func detachChecksum(data []byte) ([]byte, uint32) {
	msgLen := len(data)
	msg := data[:msgLen-crc32.Size]
	checksum := binary.LittleEndian.Uint32(data[msgLen-crc32.Size:])

	return msg, checksum
}

// calculateChecksum returns the crc32c value of the message
func calculateChecksum(msg []byte) uint32 {
	table := crc32.MakeTable(crc32.Castagnoli)
	return crc32.Checksum(msg, table)
}
