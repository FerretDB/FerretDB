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
	"encoding/json"
	"fmt"
	"io"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/types/fjson"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// OpMsgSection is one or more sections contained in an OpMsg.
type OpMsgSection struct {
	Kind       byte
	Identifier string
	Documents  []*types.Document // TODO https://github.com/FerretDB/FerretDB/issues/274
}

// OpMsg is an extensible message format designed to subsume the functionality of other opcodes.
type OpMsg struct {
	FlagBits OpMsgFlags
	Checksum uint32

	sections []OpMsgSection
}

// SetSections of the OpMsg.
func (msg *OpMsg) SetSections(sections ...OpMsgSection) error {
	msg.sections = sections
	_, err := msg.Document()
	if err != nil {
		return lazyerrors.Error(err)
	}
	return nil
}

// Document returns the value of msg as a types.Document.
func (msg *OpMsg) Document() (*types.Document, error) {
	var doc *types.Document

	for _, section := range msg.sections {
		switch section.Kind {
		case 0:
			if l := len(section.Documents); l != 1 {
				return nil, lazyerrors.Errorf("wire.OpMsg.Document: %d documents in kind 0 section", l)
			}
			if doc != nil {
				return nil, lazyerrors.Errorf("wire.OpMsg.Document: doc is not empty already: %+v", doc)
			}

			// do a shallow copy of the document that we would modify if there are kind 1 sections
			doc = must.NotFail(types.NewDocument())
			d := section.Documents[0]
			m := d.Map()
			for _, k := range d.Keys() {
				doc.Set(k, m[k])
			}

		case 1:
			if section.Identifier == "" {
				return nil, lazyerrors.New("wire.OpMsg.Document: empty section identifier")
			}
			if doc == nil {
				return nil, lazyerrors.New("wire.OpMsg.Document: doc is empty")
			}

			m := doc.Map()
			if _, ok := m[section.Identifier]; ok {
				return nil, lazyerrors.Errorf("wire.OpMsg.Document: doc already has %q key", section.Identifier)
			}

			a := types.MakeArray(len(section.Documents)) // may be zero
			for _, d := range section.Documents {
				if err := a.Append(d); err != nil {
					return nil, lazyerrors.Error(err)
				}
			}

			doc.Set(section.Identifier, a)

		default:
			return nil, lazyerrors.Errorf("wire.OpMsg.Document: unknown kind %d", section.Kind)
		}
	}

	return doc, nil
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

			d, err := types.ConvertDocument(&doc)
			if err != nil {
				return lazyerrors.Error(err)
			}
			section.Documents = []*types.Document{d}

		case 1:
			var secSize int32
			if err := binary.Read(bufr, binary.LittleEndian, &secSize); err != nil {
				return lazyerrors.Error(err)
			}

			if secSize < 5 {
				return lazyerrors.Errorf("wire.OpMsg.readFrom: invalid kind 1 section length %d", secSize)
			}

			sec := make([]byte, secSize-4)
			if n, err := io.ReadFull(bufr, sec); err != nil {
				return lazyerrors.Errorf("expected %d, read %d: %w", len(sec), n, err)
			}

			secr := bufio.NewReader(bytes.NewReader(sec))

			var id bson.CString
			if err := id.ReadFrom(secr); err != nil {
				return lazyerrors.Error(err)
			}
			section.Identifier = string(id)

			for {
				if _, err := secr.Peek(1); err == io.EOF {
					break
				}

				var doc bson.Document
				if err := doc.ReadFrom(secr); err != nil {
					return lazyerrors.Error(err)
				}

				d, err := types.ConvertDocument(&doc)
				if err != nil {
					return lazyerrors.Error(err)
				}
				section.Documents = append(section.Documents, d)
			}

		default:
			return lazyerrors.Errorf("kind is %d", section.Kind)
		}

		msg.sections = append(msg.sections, section)

		peekBytes := 1
		if msg.FlagBits.FlagSet(OpMsgChecksumPresent) {
			peekBytes = 5
		}
		if _, err := bufr.Peek(peekBytes); err == io.EOF {
			break
		}
	}

	if msg.FlagBits.FlagSet(OpMsgChecksumPresent) {
		if err := binary.Read(bufr, binary.LittleEndian, &msg.Checksum); err != nil {
			return lazyerrors.Error(err)
		}
	}

	if _, err := msg.Document(); err != nil {
		return lazyerrors.Error(err)
	}

	// TODO validate checksum

	return nil
}

// UnmarshalBinary reads an OpMsg from a byte array.
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

// MarshalBinary writes an OpMsg to a byte array.
func (msg *OpMsg) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	if err := binary.Write(bufw, binary.LittleEndian, msg.FlagBits); err != nil {
		return nil, lazyerrors.Error(err)
	}

	for _, section := range msg.sections {
		if err := bufw.WriteByte(section.Kind); err != nil {
			return nil, lazyerrors.Error(err)
		}

		switch section.Kind {
		case 0:
			if l := len(section.Documents); l != 1 {
				panic(fmt.Sprintf("%d documents in section with kind 0", l))
			}

			d, err := bson.ConvertDocument(section.Documents[0])
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			if d.WriteTo(bufw); err != nil {
				return nil, lazyerrors.Error(err)
			}

		case 1:
			var secBuf bytes.Buffer
			secw := bufio.NewWriter(&secBuf)

			if err := bson.CString(section.Identifier).WriteTo(secw); err != nil {
				return nil, lazyerrors.Error(err)
			}

			for _, doc := range section.Documents {
				d, err := bson.ConvertDocument(doc)
				if err != nil {
					return nil, lazyerrors.Error(err)
				}
				if d.WriteTo(secw); err != nil {
					return nil, lazyerrors.Error(err)
				}
			}

			if err := secw.Flush(); err != nil {
				return nil, lazyerrors.Error(err)
			}

			if err := binary.Write(bufw, binary.LittleEndian, int32(secBuf.Len()+4)); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if _, err := bufw.Write(secBuf.Bytes()); err != nil {
				return nil, lazyerrors.Error(err)
			}

		default:
			return nil, lazyerrors.Errorf("kind is %d", section.Kind)
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

// String returns a string representation for logging.
func (msg *OpMsg) String() string {
	if msg == nil {
		return "<nil>"
	}

	m := map[string]any{
		"FlagBits": msg.FlagBits,
		"Checksum": msg.Checksum,
	}

	sections := make([]map[string]any, len(msg.sections))
	for i, section := range msg.sections {
		s := map[string]any{
			"Kind": section.Kind,
		}
		switch section.Kind {
		case 0:
			b := must.NotFail(fjson.Marshal(section.Documents[0]))
			s["Document"] = json.RawMessage(b)
		case 1:
			s["Identifier"] = section.Identifier
			docs := make([]json.RawMessage, len(section.Documents))
			for j, d := range section.Documents {
				b := must.NotFail(fjson.Marshal(d))
				docs[j] = json.RawMessage(b)
			}
			s["Documents"] = docs
		}

		sections[i] = s
	}

	m["Sections"] = sections

	return string(must.NotFail(json.MarshalIndent(m, "", "  ")))
}

// check interfaces
var (
	_ MsgBody = (*OpMsg)(nil)
)
