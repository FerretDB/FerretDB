// Copyright 2021 Baltoro OÜ.
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
	"encoding/json"
	"io"

	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

//go:generate ../../bin/stringer -linecomment -type OpCode

type OpCode int32

const (
	OP_REPLY = OpCode(1)    // OP_REPLY
	OP_MSG   = OpCode(2013) // OP_MSG
	OP_QUERY = OpCode(2004) // OP_QUERY
)

func (i OpCode) MarshalJSON() ([]byte, error) {
	return []byte(`"` + i.String() + `"`), nil
}

type MsgHeader struct {
	MessageLength int32
	RequestID     int32
	ResponseTo    int32
	OpCode        OpCode
}

const (
	MsgHeaderLen = 16
	MaxMsgLen    = 48000000
)

func (msg *MsgHeader) readFrom(r *bufio.Reader) error {
	b := make([]byte, MsgHeaderLen)
	if n, err := io.ReadFull(r, b); err != nil {
		if err == io.EOF {
			return err
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

func (msg *MsgHeader) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, uint32(msg.MessageLength))
	binary.Write(&buf, binary.LittleEndian, uint32(msg.RequestID))
	binary.Write(&buf, binary.LittleEndian, uint32(msg.ResponseTo))
	binary.Write(&buf, binary.LittleEndian, uint32(msg.OpCode))

	return buf.Bytes(), nil
}

// check interfaces
var (
	_ json.Marshaler = OpCode(0)
)
