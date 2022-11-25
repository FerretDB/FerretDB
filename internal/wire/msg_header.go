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
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

//go:generate ../../bin/stringer -linecomment -type OpCode

// OpCode represents wire operation code.
type OpCode int32

const (
	// OpCodeReply is used for the initial handshake with old clients.
	// It is not used otherwise and is deprecated.
	OpCodeReply = OpCode(1) // OP_REPLY

	// OpCodeUpdate is deprecated.
	OpCodeUpdate = OpCode(2001) // OP_UPDATE

	// OpCodeInsert is deprecated and unused.
	OpCodeInsert = OpCode(2002) // OP_INSERT

	// OpCodeGetByOID is deprecated and unused.
	OpCodeGetByOID = OpCode(2003) // OP_GET_BY_OID

	// OpCodeQuery is used for the initial handshake with old clients.
	// It is not used otherwise and is deprecated.
	OpCodeQuery = OpCode(2004) // OP_QUERY

	// OpCodeGetMore is deprecated and unused.
	OpCodeGetMore = OpCode(2005) // OP_GET_MORE

	// OpCodeDelete is deprecated and unused.
	OpCodeDelete = OpCode(2006) // OP_DELETE

	// OpCodeKillCursors is deprecated and unused.
	OpCodeKillCursors = OpCode(2007) // OP_KILL_CURSORS

	// OpCodeCompressed is not implemented yet.
	OpCodeCompressed = OpCode(2012) // OP_COMPRESSED

	// OpCodeMsg is the main operation for client-server communication.
	OpCodeMsg = OpCode(2013) // OP_MSG
)

// MsgHeader in general, each message consists of a standard message header followed by request-specific data.
type MsgHeader struct {
	MessageLength int32
	RequestID     int32
	ResponseTo    int32
	OpCode        OpCode
}

const (
	// MsgHeaderLen is an expected len of the header.
	MsgHeaderLen = 16

	// MaxMsgLen is the maximum message length.
	MaxMsgLen = 48000000
)

// readFrom reads header.
//
// Error is ErrZeroRead if zero bytes was read.
func (msg *MsgHeader) readFrom(r *bufio.Reader) error {
	b := make([]byte, MsgHeaderLen)
	if n, err := io.ReadFull(r, b); err != nil {
		if err == io.EOF {
			return ErrZeroRead
		}
		return lazyerrors.Errorf("expected %d, read %d: %w", len(b), n, err)
	}

	msg.MessageLength = int32(binary.LittleEndian.Uint32(b[0:4]))
	msg.RequestID = int32(binary.LittleEndian.Uint32(b[4:8]))
	msg.ResponseTo = int32(binary.LittleEndian.Uint32(b[8:12]))
	msg.OpCode = OpCode(binary.LittleEndian.Uint32(b[12:16]))

	if msg.MessageLength < MsgHeaderLen || msg.MessageLength > MaxMsgLen {
		return lazyerrors.Errorf("invalid message length %d", msg.MessageLength)
	}

	return nil
}

func (msg *MsgHeader) writeTo(w *bufio.Writer) error {
	v, err := msg.MarshalBinary()
	if err != nil {
		return lazyerrors.Error(err)
	}

	if _, err := w.Write(v); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// MarshalBinary writes a MsgHeader to a byte array.
func (msg *MsgHeader) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, uint32(msg.MessageLength))
	binary.Write(&buf, binary.LittleEndian, uint32(msg.RequestID))
	binary.Write(&buf, binary.LittleEndian, uint32(msg.ResponseTo))
	binary.Write(&buf, binary.LittleEndian, uint32(msg.OpCode))

	return buf.Bytes(), nil
}

// String returns a string representation for logging.
func (msg *MsgHeader) String() string {
	if msg == nil {
		return "<nil>"
	}

	return fmt.Sprintf(
		"length: %5d, id: %4d, response_to: %4d, opcode: %s",
		msg.MessageLength, msg.RequestID, msg.ResponseTo, msg.OpCode,
	)
}

// check interfaces
var (
	_ fmt.Stringer = OpCode(0)
	_ fmt.Stringer = (*MsgHeader)(nil)
)
