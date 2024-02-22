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
	"encoding/json"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/bson2"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/types/fjson"
	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// OpMsgSection is one or more sections contained in an OpMsg.
type OpMsgSection struct {
	Kind       byte
	Identifier string
	documents  []bson2.RawDocument
}

// MakeOpMsgSection creates [OpMsgSection] with a single document.
func MakeOpMsgSection(doc *types.Document) OpMsgSection {
	raw := must.NotFail(must.NotFail(bson2.ConvertDocument(doc)).Encode())
	return OpMsgSection{
		documents: []bson2.RawDocument{raw},
	}
}

// RawDocuments returns raw documents of the section.
func (s *OpMsgSection) RawDocuments() []bson2.RawDocument {
	return s.documents
}

// OpMsg is the main wire protocol message type.
type OpMsg struct {
	Flags OpMsgFlags

	sections []OpMsgSection
	checksum uint32
}

// Sections returns the sections of the OpMsg.
func (msg *OpMsg) Sections() []OpMsgSection {
	return msg.sections
}

// SetSections sets sections of the OpMsg.
func (msg *OpMsg) SetSections(sections ...OpMsgSection) error {
	msg.sections = sections
	_, err := msg.Document()
	if err != nil {
		return lazyerrors.Error(err)
	}
	return nil
}

// Document returns the value of msg as a [types.Document].
//
// All sections are merged together.
func (msg *OpMsg) Document() (*types.Document, error) {
	// Sections of kind 1 may come before the section of kind 0,
	// but the command is defined by the first key in the section of kind 0.
	// Reorder documents to set keys in the right order.

	docs := make([]*types.Document, 0, len(msg.sections))

	for _, section := range msg.sections {
		if section.Kind != 0 {
			continue
		}

		if l := len(section.documents); l != 1 {
			return nil, lazyerrors.Errorf("wire.OpMsg.Document: %d documents in kind 0 section", l)
		}

		doc, err := section.documents[0].Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		docs = append(docs, doc)
	}

	for _, section := range msg.sections {
		if section.Kind == 0 {
			continue
		}

		if section.Kind != 1 {
			panic(fmt.Sprintf("unknown kind %d", section.Kind))
		}

		if section.Identifier == "" {
			return nil, lazyerrors.New("wire.OpMsg.Document: empty section identifier")
		}

		a := types.MakeArray(len(section.documents))
		for _, d := range section.documents {
			doc, err := d.Convert()
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			a.Append(doc)
		}

		docs = append(docs, must.NotFail(types.NewDocument(section.Identifier, a)))
	}

	res := types.MakeDocument(2)

	for _, doc := range docs {
		values := doc.Values()

		for i, k := range doc.Keys() {
			res.Set(k, values[i])
		}
	}

	return res, nil
}

func (msg *OpMsg) msgbody() {}

// check implements [MsgBody] interface.
func (msg *OpMsg) check() error {
	for _, s := range msg.sections {
		for _, d := range s.documents {
			if _, err := d.DecodeDeep(); err != nil {
				lazyerrors.Error(err)
			}
		}
	}

	return nil
}

// UnmarshalBinaryNocopy implements [MsgBody] interface.
func (msg *OpMsg) UnmarshalBinaryNocopy(b []byte) error {
	if len(b) < 6 {
		return lazyerrors.Errorf("len=%d", len(b))
	}

	msg.Flags = OpMsgFlags(binary.LittleEndian.Uint32(b[0:4]))

	offset := 4
	for {
		var section OpMsgSection
		section.Kind = b[offset]
		offset++

		switch section.Kind {
		case 0:
			l, err := bson2.FindRaw(b[offset:])
			if err != nil {
				return lazyerrors.Error(err)
			}

			section.documents = []bson2.RawDocument{b[offset : offset+l]}
			offset += l

		case 1:
			if len(b) < offset+4 {
				return lazyerrors.Errorf("len(b) = %d, offset = %d", len(b), offset)
			}

			secSize := int(binary.LittleEndian.Uint32(b[offset:offset+4])) - 4
			if secSize < 5 {
				return lazyerrors.Errorf("size = %d", secSize)
			}

			offset += 4

			var err error

			if len(b) < offset {
				return lazyerrors.Errorf("len(b) = %d, offset = %d", len(b), offset)
			}

			section.Identifier, err = bson2.DecodeCString(b[offset:])
			if err != nil {
				return lazyerrors.Error(err)
			}

			offset += bson2.SizeCString(section.Identifier)
			secSize -= bson2.SizeCString(section.Identifier)

			for secSize != 0 {
				if secSize < 0 {
					return lazyerrors.Errorf("size = %d", secSize)
				}

				if len(b) < offset {
					return lazyerrors.Errorf("len(b) = %d, offset = %d", len(b), offset)
				}

				l, err := bson2.FindRaw(b[offset:])
				if err != nil {
					return lazyerrors.Error(err)
				}

				section.documents = append(section.documents, b[offset:offset+l])
				offset += l
				secSize -= l
			}

		default:
			return lazyerrors.Errorf("kind is %d", section.Kind)
		}

		msg.sections = append(msg.sections, section)

		if msg.Flags.FlagSet(OpMsgChecksumPresent) {
			if offset == len(b)-4 {
				break
			}
		} else {
			if offset == len(b) {
				break
			}
		}
	}

	if msg.Flags.FlagSet(OpMsgChecksumPresent) {
		// Move checksum validation here. It needs header data to be available.
		// TODO https://github.com/FerretDB/FerretDB/issues/2690

		msg.checksum = binary.LittleEndian.Uint32(b[offset:])
	}

	if debugbuild.Enabled {
		if err := msg.check(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// MarshalBinary writes an OpMsg to a byte array.
func (msg *OpMsg) MarshalBinary() ([]byte, error) {
	if debugbuild.Enabled {
		if err := msg.check(); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	b := make([]byte, 4, 16)

	binary.LittleEndian.PutUint32(b, uint32(msg.Flags))

	for _, section := range msg.sections {
		b = append(b, section.Kind)

		switch section.Kind {
		case 0:
			if l := len(section.documents); l != 1 {
				panic(fmt.Sprintf("%d documents in section with kind 0", l))
			}

			b = append(b, section.documents[0]...)

		case 1:
			sec := make([]byte, bson2.SizeCString(section.Identifier))
			bson2.EncodeCString(sec, section.Identifier)

			for _, doc := range section.documents {
				sec = append(sec, doc...)
			}

			var size [4]byte
			binary.LittleEndian.PutUint32(size[:], uint32(len(sec)+4))
			b = append(b, size[:]...)
			b = append(b, sec...)

		default:
			return nil, lazyerrors.Errorf("kind is %d", section.Kind)
		}
	}

	if msg.Flags.FlagSet(OpMsgChecksumPresent) {
		// Calculate checksum before writing it. It needs header data to be ready and available here.
		// TODO https://github.com/FerretDB/FerretDB/issues/2690

		var checksum [4]byte
		binary.LittleEndian.PutUint32(checksum[:], msg.checksum)
		b = append(b, checksum[:]...)
	}

	return b, nil
}

// String returns a string representation for logging.
func (msg *OpMsg) String() string {
	if msg == nil {
		return "<nil>"
	}

	m := map[string]any{
		"FlagBits": msg.Flags,
		"Checksum": msg.checksum,
	}

	sections := make([]map[string]any, len(msg.sections))
	for i, section := range msg.sections {
		s := map[string]any{
			"Kind": section.Kind,
		}
		switch section.Kind {
		case 0:
			doc := must.NotFail(section.documents[0].Convert())
			b := must.NotFail(fjson.Marshal(doc))
			s["Document"] = json.RawMessage(b)

		case 1:
			s["Identifier"] = section.Identifier
			docs := make([]json.RawMessage, len(section.documents))

			for j, d := range section.documents {
				doc := must.NotFail(d.Convert())
				b := must.NotFail(fjson.Marshal(doc))
				docs[j] = json.RawMessage(b)
			}

			s["Documents"] = docs

		default:
			panic(fmt.Sprintf("unknown kind %d", section.Kind))
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
