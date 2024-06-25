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

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// typesValidation, when true, enables validation of types in wire messages.
const typesValidation = true

// MakeOpMsgSection creates [opMsgSection] with a single document.
func MakeOpMsgSection(doc *types.Document) opMsgSection {
	raw := must.NotFail(must.NotFail(bson.ConvertDocument(doc)).Encode())

	return opMsgSection{
		documents: []bson.RawDocument{raw},
	}
}

// OpMsg is the main wire protocol message type.
type OpMsg struct {
	// The order of fields is weird to make the struct smaller due to alignment.
	// The wire order is: flags, sections, optional checksum.

	sections []opMsgSection
	Flags    OpMsgFlags
	checksum uint32
}

// NewOpMsg creates a message with a single section of kind 0 with a single raw document.
func NewOpMsg(doc bson.AnyDocument) (*OpMsg, error) {
	raw, err := doc.Encode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var msg OpMsg
	if err = msg.SetSections(opMsgSection{documents: []bson.RawDocument{raw}}); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &msg, nil
}

// Sections returns the sections of the OpMsg.
func (msg *OpMsg) Sections() []opMsgSection {
	return msg.sections
}

// SetSections sets sections of the OpMsg.
func (msg *OpMsg) SetSections(sections ...opMsgSection) error {
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
		if section.kind != 0 {
			continue
		}

		doc, err := section.documents[0].Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		docs = append(docs, doc)
	}

	for _, section := range msg.sections {
		if section.kind == 0 {
			continue
		}

		a := types.MakeArray(len(section.documents))

		for _, d := range section.documents {
			doc, err := d.Convert()
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			a.Append(doc)
		}

		docs = append(docs, must.NotFail(types.NewDocument(section.identifier, a)))
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
func (msg *OpMsg) RawSections() (bson.RawDocument, []byte) {
	var spec bson.RawDocument
	var seq []byte

	for _, s := range msg.Sections() {
		switch s.kind {
		case 0:
			spec = s.documents[0]

		case 1:
			for _, d := range s.documents {
				seq = append(seq, d...)
			}
		}
	}

	return spec, seq
}

// RawDocument returns the value of msg as a [bson.RawDocument].
//
// The error is returned if msg contains anything other than a single section of kind 0
// with a single document.
func (msg *OpMsg) RawDocument() (bson.RawDocument, error) {
	if err := checkSections(msg.sections); err != nil {
		return nil, err
	}

	s := msg.sections[0]
	if s.kind != 0 || s.identifier != "" {
		return nil, lazyerrors.Errorf(`expected section 0/"", got %d/%q`, s.kind, s.identifier)
	}

	return s.documents[0], nil
}

func (msg *OpMsg) msgbody() {}

// check implements [MsgBody] interface.
func (msg *OpMsg) check() error {
	for _, s := range msg.sections {
		for _, d := range s.documents {
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
		var section opMsgSection
		section.kind = b[offset]
		offset++

		switch section.kind {
		case 0:
			l, err := bson.FindRaw(b[offset:])
			if err != nil {
				return lazyerrors.Error(err)
			}

			section.documents = []bson.RawDocument{b[offset : offset+l]}
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

			section.identifier, err = bson.DecodeCString(b[offset:])
			if err != nil {
				return lazyerrors.Error(err)
			}

			offset += bson.SizeCString(section.identifier)
			secSize -= bson.SizeCString(section.identifier)

			for secSize != 0 {
				if secSize < 0 {
					return lazyerrors.Errorf("size = %d", secSize)
				}

				if len(b) < offset {
					return lazyerrors.Errorf("len(b) = %d, offset = %d", len(b), offset)
				}

				l, err := bson.FindRaw(b[offset:])
				if err != nil {
					return lazyerrors.Error(err)
				}

				section.documents = append(section.documents, b[offset:offset+l])
				offset += l
				secSize -= l
			}

		default:
			return lazyerrors.Errorf("kind is %d", section.kind)
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
		b = append(b, section.kind)

		switch section.kind {
		case 0:
			b = append(b, section.documents[0]...)

		case 1:
			sec := make([]byte, bson.SizeCString(section.identifier))
			bson.EncodeCString(sec, section.identifier)

			for _, doc := range section.documents {
				sec = append(sec, doc...)
			}

			var size [4]byte
			binary.LittleEndian.PutUint32(size[:], uint32(len(sec)+4))
			b = append(b, size[:]...)
			b = append(b, sec...)

		default:
			return nil, lazyerrors.Errorf("kind is %d", section.kind)
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

// logMessage returns a string representation for logging.
func (msg *OpMsg) logMessage(logFunc func(v any) string) string {
	if msg == nil {
		return "<nil>"
	}

	m := must.NotFail(bson.NewDocument(
		"FlagBits", msg.Flags.String(),
		"Checksum", int64(msg.checksum),
	))

	sections := bson.MakeArray(len(msg.sections))
	for _, section := range msg.sections {
		s := must.NotFail(bson.NewDocument(
			"Kind", int32(section.kind),
		))

		switch section.kind {
		case 0:
			doc, err := section.documents[0].DecodeDeep()
			if err == nil {
				must.NoError(s.Add("Document", doc))
			} else {
				must.NoError(s.Add("DocumentError", err.Error()))
			}

		case 1:
			must.NoError(s.Add("Identifier", section.identifier))
			docs := bson.MakeArray(len(section.documents))

			for _, d := range section.documents {
				doc, err := d.DecodeDeep()
				if err == nil {
					must.NoError(docs.Add(doc))
				} else {
					must.NoError(docs.Add(must.NotFail(bson.NewDocument("error", err.Error()))))
				}
			}

			must.NoError(s.Add("Documents", docs))

		default:
			panic(fmt.Sprintf("unknown kind %d", section.kind))
		}

		must.NoError(sections.Add(s))
	}

	must.NoError(m.Add("Sections", sections))

	return logFunc(m)
}

// String returns a string representation for logging.
func (msg *OpMsg) String() string {
	return msg.logMessage(bson.LogMessage)
}

// StringBlock returns an indented string representation for logging.
func (msg *OpMsg) StringBlock() string {
	return msg.logMessage(bson.LogMessageBlock)
}

// StringFlow returns an unindented string representation for logging.
func (msg *OpMsg) StringFlow() string {
	return msg.logMessage(bson.LogMessageFlow)
}

// check interfaces
var (
	_ MsgBody = (*OpMsg)(nil)
)
