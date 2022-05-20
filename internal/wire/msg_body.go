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
	"fmt"
	"io"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

type MsgBody interface {
	readFrom(*bufio.Reader) error
	encoding.BinaryUnmarshaler
	encoding.BinaryMarshaler
	fmt.Stringer

	msgbody() // seal for go-sumtype
}

//go-sumtype:decl MsgBody

func ReadMessage(r *bufio.Reader) (*MsgHeader, MsgBody, error) {
	var header MsgHeader
	if err := header.readFrom(r); err != nil {
		if err == io.EOF {
			return nil, nil, err
		}
		return nil, nil, lazyerrors.Error(err)
	}

	b := make([]byte, header.MessageLength-MsgHeaderLen)
	if n, err := io.ReadFull(r, b); err != nil {
		return nil, nil, lazyerrors.Errorf("expected %d, read %d: %w", len(b), n, err)
	}

	switch header.OpCode {
	case OpCodeReply:
		var reply OpReply
		if err := reply.UnmarshalBinary(b); err != nil {
			return nil, nil, lazyerrors.Error(err)
		}

		return &header, &reply, nil

	case OpCodeMsg:
		var msg OpMsg
		if err := msg.UnmarshalBinary(b); err != nil {
			return nil, nil, lazyerrors.Error(err)
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
		fallthrough

	default:
		return nil, nil, lazyerrors.Errorf("unhandled opcode %s", header.OpCode)
	}
}

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
