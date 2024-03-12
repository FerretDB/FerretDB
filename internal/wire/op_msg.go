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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/bson2"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// typesValidation, when true, enables validation of types in wire messages.
const typesValidation = true

// OpMsgSection is one or more sections contained in an OpMsg.
type OpMsgSection struct {
	// The order of fields is weird to make the struct smaller due to alignment.
	// The wire order is: kind, identifier, documents.

	Identifier string
	Documents  []bson2.RawDocument
	Kind       byte
}

// MakeOpMsgSection creates [OpMsgSection] with a single document.
func MakeOpMsgSection(doc *types.Document) OpMsgSection {
	raw := must.NotFail(must.NotFail(bson2.ConvertDocument(doc)).Encode())

	return OpMsgSection{
		Documents: []bson2.RawDocument{raw},
	}
}

// OpMsg is the main wire protocol message type.
type OpMsg struct {
	// The order of fields is weird to make the struct smaller due to alignment.
	// The wire order is: flags, sections, optional checksum.

	sections []OpMsgSection
	Flags    OpMsgFlags
	checksum uint32
}

// NewOpMsg creates a message with a single section of kind 0 with a single raw document.
func NewOpMsg(raw bson2.RawDocument) (*OpMsg, error) {
	var msg OpMsg
	if err := msg.SetSections(OpMsgSection{Documents: []bson2.RawDocument{raw}}); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &msg, nil
}

// checkSections checks given sections.
func checkSections(sections []OpMsgSection) error {
	if len(sections) == 0 {
		return lazyerrors.New("no sections")
	}

	var kind0Found bool

	for _, s := range sections {
		switch s.Kind {
		case 0:
			if kind0Found {
				return lazyerrors.New("multiple kind 0 sections")
			}
			kind0Found = true

			if s.Identifier != "" {
				return lazyerrors.New("kind 0 section has identifier")
			}

			if len(s.Documents) != 1 {
				return lazyerrors.Errorf("kind 0 section has %d documents", len(s.Documents))
			}

		case 1:
			if s.Identifier == "" {
				return lazyerrors.New("kind 1 section has no identifier")
			}

		default:
			return lazyerrors.Errorf("unknown kind %d", s.Kind)
		}
	}

	return nil
}

// Sections returns the sections of the OpMsg.
func (msg *OpMsg) Sections() []OpMsgSection {
	return msg.sections
}

// SetSections sets sections of the OpMsg.
func (msg *OpMsg) SetSections(sections ...OpMsgSection) error {
	if err := checkSections(sections); err != nil {
		return lazyerrors.Error(err)
	}

	msg.sections = sections

	if debugbuild.Enabled {
		if err := msg.check(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	if typesValidation {
		if _, err := msg.Document(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// Document returns the value of msg as a [types.Document].
//
// All sections are merged together.
func (msg *OpMsg) Document() (*types.Document, error) {
	if err := checkSections(msg.sections); err != nil {
		return nil, lazyerrors.Error(err)
	}

	docs := make([]*types.Document, 0, len(msg.sections))

	// Sections of kind 1 may come before the section of kind 0,
	// but the command is defined by the first key in the section of kind 0.
	// Reorder documents to set keys in the right order.

	for _, section := range msg.sections {
		if section.Kind != 0 {
			continue
		}

		doc, err := section.Documents[0].Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		docs = append(docs, doc)
	}

	for _, section := range msg.sections {
		if section.Kind == 0 {
			continue
		}

		a := types.MakeArray(len(section.Documents))

		for _, d := range section.Documents {
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

	if err := validateValue(res); err != nil {
		res.Remove("lsid") // to simplify error message
		return nil, newValidationError(fmt.Errorf("wire.OpMsg.Document: validation failed for %v with: %v",
			types.FormatAnyValue(res),
			err,
		))
	}

	return res, nil
}

// RawSections returns the value of section with kind 0 and the value of all sections with kind 1.
func (msg *OpMsg) RawSections() (bson2.RawDocument, []byte) {
	var spec bson2.RawDocument
	var seq []byte

	for _, s := range msg.Sections() {
		switch s.Kind {
		case 0:
			spec = s.Documents[0]

		case 1:
			for _, d := range s.Documents {
				seq = append(seq, d...)
			}
		}
	}

	return spec, seq
}

// RawDocument returns the value of msg as a [bson2.RawDocument].
//
// The error is returned if msg contains anything other than a single section of kind 0
// with a single document.
func (msg *OpMsg) RawDocument() (bson2.RawDocument, error) {
	if err := checkSections(msg.sections); err != nil {
		return nil, err
	}

	s := msg.sections[0]
	if s.Kind != 0 || s.Identifier != "" {
		return nil, lazyerrors.Errorf(`expected section 0/"", got %d/%q`, s.Kind, s.Identifier)
	}

	return s.Documents[0], nil
}

func (msg *OpMsg) msgbody() {}

// check implements [MsgBody] interface.
func (msg *OpMsg) check() error {
	for _, s := range msg.sections {
		for _, d := range s.Documents {
			if _, err := d.DecodeDeep(); err != nil {
				return lazyerrors.Error(err)
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

			section.Documents = []bson2.RawDocument{b[offset : offset+l]}
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

				section.Documents = append(section.Documents, b[offset:offset+l])
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

	if err := checkSections(msg.sections); err != nil {
		return lazyerrors.Error(err)
	}

	if debugbuild.Enabled {
		if err := msg.check(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	if typesValidation {
		if _, err := msg.Document(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// MarshalBinary writes an OpMsg to a byte array.
func (msg *OpMsg) MarshalBinary() ([]byte, error) {
	if err := checkSections(msg.sections); err != nil {
		return nil, lazyerrors.Error(err)
	}

	if debugbuild.Enabled {
		if err := msg.check(); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	if typesValidation {
		if _, err := msg.Document(); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	b := make([]byte, 4, 16)

	binary.LittleEndian.PutUint32(b, uint32(msg.Flags))

	for _, section := range msg.sections {
		b = append(b, section.Kind)

		switch section.Kind {
		case 0:
			b = append(b, section.Documents[0]...)

		case 1:
			sec := make([]byte, bson2.SizeCString(section.Identifier))
			bson2.EncodeCString(sec, section.Identifier)

			for _, doc := range section.Documents {
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

	m := must.NotFail(bson2.NewDocument(
		"FlagBits", msg.Flags.String(),
		"Checksum", int64(msg.checksum),
	))

	sections := bson2.MakeArray(len(msg.sections))
	for _, section := range msg.sections {
		s := must.NotFail(bson2.NewDocument(
			"Kind", int32(section.Kind),
		))

		switch section.Kind {
		case 0:
			doc, err := section.Documents[0].DecodeDeep()
			if err == nil {
				must.NoError(s.Add("Document", doc))
			} else {
				must.NoError(s.Add("DocumentError", err.Error()))
			}

		case 1:
			must.NoError(s.Add("Identifier", section.Identifier))
			docs := bson2.MakeArray(len(section.Documents))

			for _, d := range section.Documents {
				doc, err := d.DecodeDeep()
				if err == nil {
					must.NoError(docs.Add(doc))
				} else {
					must.NoError(docs.Add(must.NotFail(bson2.NewDocument("error", err.Error()))))
				}
			}

			must.NoError(s.Add("Documents", docs))

		default:
			panic(fmt.Sprintf("unknown kind %d", section.Kind))
		}

		must.NoError(sections.Add(s))
	}

	must.NoError(m.Add("Sections", sections))

	return m.LogMessage()
}

// check interfaces
var (
	_ MsgBody = (*OpMsg)(nil)
)
