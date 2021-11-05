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

	"github.com/MangoDB-io/MangoDB/internal/bson"
	"github.com/MangoDB-io/MangoDB/internal/types"
	lazyerrors "github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

type OpMsgSection struct {
	Kind       byte
	Identifier string
	Documents  []types.Document
}

type OpMsg struct {
	FlagBits OpMsgFlags
	// Deprecated: remove.
	Documents []types.Document
	Sections  []OpMsgSection
	Checksum  uint32
}

func (msg *OpMsg) msgbody() {}

func (msg *OpMsg) readFrom(bufr *bufio.Reader) error {
	if err := binary.Read(bufr, binary.LittleEndian, &msg.FlagBits); err != nil {
		return lazyerrors.Error(err)
	}

	for {
		var section OpMsgSection
		if err := binary.Read(bufr, binary.LittleEndian, &section.Kind); err != nil {
			return lazyerrors.Error(err)
		}

		switch section.Kind {
		case 0:
			var doc bson.Document
			if err := doc.ReadFrom(bufr); err != nil {
				return lazyerrors.Error(err)
			}

			d := types.MustNewDocument(&doc)
			section.Documents = append(section.Documents, d)
			msg.Documents = append(msg.Documents, d)

		case 1:
			var seqSize int32
			if err := binary.Read(bufr, binary.LittleEndian, &seqSize); err != nil {
				return lazyerrors.Error(err)
			}

			seq := make([]byte, seqSize-4)
			if n, err := io.ReadFull(bufr, seq); err != nil {
				return lazyerrors.Errorf("expected %d, read %d: %w", len(seq), n, err)
			}

			seqR := bufio.NewReader(bytes.NewReader(seq))

			var id bson.CString
			if err := id.ReadFrom(seqR); err != nil {
				return lazyerrors.Error(err)
			}
			section.Identifier = string(id)

			for {
				if _, err := seqR.Peek(1); err == io.EOF {
					break
				}

				var doc bson.Document
				if err := doc.ReadFrom(seqR); err != nil {
					return lazyerrors.Error(err)
				}

				d := types.MustNewDocument(&doc)
				section.Documents = append(section.Documents, d)
				msg.Documents = append(msg.Documents, d)
			}

		default:
			return lazyerrors.Errorf("kind is %d", section.Kind)
		}

		msg.Sections = append(msg.Sections, section)

		peekBytes := 1
		if msg.FlagBits.FlagSet(OpMsgChecksumPresent) {
			peekBytes = 5
		}

		if _, err := bufr.Peek(peekBytes); err == io.EOF {
			break
		}
	}

	// TODO
	msg.Sections = nil

	if msg.FlagBits.FlagSet(OpMsgChecksumPresent) {
		if err := binary.Read(bufr, binary.LittleEndian, &msg.Checksum); err != nil {
			return lazyerrors.Error(err)
		}
	}

	// TODO validate checksum

	return nil
}

func (msg *OpMsg) UnmarshalBinary(b []byte) error {
	br := bytes.NewReader(b)
	bufr := bufio.NewReader(br)

	if err := msg.readFrom(bufr); err != nil {
		return lazyerrors.Error(err)
	}

	if _, err := bufr.Peek(1); err != io.EOF {
		return lazyerrors.Errorf("unexpected end of the OpMsg: %v", err)
	}

	return nil
}

func (msg *OpMsg) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	if err := binary.Write(bufw, binary.LittleEndian, msg.FlagBits); err != nil {
		return nil, lazyerrors.Error(err)
	}

	for _, doc := range msg.Documents {
		// kind
		if err := bufw.WriteByte(0); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if err := bson.MustNewDocument(doc).WriteTo(bufw); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	if msg.FlagBits.FlagSet(OpMsgChecksumPresent) {
		if err := binary.Write(bufw, binary.LittleEndian, msg.Checksum); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	if err := bufw.Flush(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return buf.Bytes(), nil
}

func (msg *OpMsg) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"FlagBits": msg.FlagBits,
		"Checksum": msg.Checksum,
	}

	docs := make([]interface{}, len(msg.Documents))
	for i, d := range msg.Documents {
		docs[i] = bson.MustNewDocument(d)
	}

	m["Documents"] = docs

	return json.Marshal(m)
}

// check interfaces
var (
	_ MsgBody = (*OpMsg)(nil)
)
