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

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// OpQuery is a deprecated request message type.
type OpQuery struct {
	// The order of fields is weird to make the struct smaller due to alignment.
	// The wire order is: flags, collection name, number to skip, number to return, query, fields selector.

	FullCollectionName   string
	query                bson.RawDocument
	returnFieldsSelector bson.RawDocument
	Flags                OpQueryFlags
	NumberToSkip         int32
	NumberToReturn       int32
}

// NewOpQuery creates a new OpQuery message.
func NewOpQuery(doc bson.AnyDocument) (*OpQuery, error) {
	raw, err := doc.Encode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &OpQuery{query: raw}, nil
}

func (query *OpQuery) msgbody() {}

// check implements [MsgBody].
func (query *OpQuery) check() error {
	if d := query.query; d != nil {
		if _, err := d.DecodeDeep(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	if s := query.returnFieldsSelector; s != nil {
		if _, err := s.DecodeDeep(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// UnmarshalBinaryNocopy implements [MsgBody].
func (query *OpQuery) UnmarshalBinaryNocopy(b []byte) error {
	if len(b) < 4 {
		return lazyerrors.Errorf("len=%d", len(b))
	}

	query.Flags = OpQueryFlags(binary.LittleEndian.Uint32(b[0:4]))

	var err error

	query.FullCollectionName, err = bson.DecodeCString(b[4:])
	if err != nil {
		return lazyerrors.Error(err)
	}

	numberLow := 4 + bson.SizeCString(query.FullCollectionName)
	if len(b) < numberLow+8 {
		return lazyerrors.Errorf("len=%d, can't unmarshal numbers", len(b))
	}

	query.NumberToSkip = int32(binary.LittleEndian.Uint32(b[numberLow : numberLow+4]))
	query.NumberToReturn = int32(binary.LittleEndian.Uint32(b[numberLow+4 : numberLow+8]))

	l, err := bson.FindRaw(b[numberLow+8:])
	if err != nil {
		return lazyerrors.Error(err)
	}
	query.query = b[numberLow+8 : numberLow+8+l]

	selectorLow := numberLow + 8 + l
	if len(b) != selectorLow {
		l, err = bson.FindRaw(b[selectorLow:])
		if err != nil {
			return lazyerrors.Error(err)
		}

		if len(b) != selectorLow+l {
			return lazyerrors.Errorf("len=%d, expected=%d", len(b), selectorLow+l)
		}
		query.returnFieldsSelector = b[selectorLow:]
	}

	if debugbuild.Enabled {
		if err := query.check(); err != nil {
			return lazyerrors.Error(err)
		}
	}

	return nil
}

// MarshalBinary implements [MsgBody].
func (query *OpQuery) MarshalBinary() ([]byte, error) {
	if debugbuild.Enabled {
		if err := query.check(); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	nameSize := bson.SizeCString(query.FullCollectionName)
	b := make([]byte, 12+nameSize+len(query.query)+len(query.returnFieldsSelector))

	binary.LittleEndian.PutUint32(b[0:4], uint32(query.Flags))

	nameHigh := 4 + nameSize
	bson.EncodeCString(b[4:nameHigh], query.FullCollectionName)

	binary.LittleEndian.PutUint32(b[nameHigh:nameHigh+4], uint32(query.NumberToSkip))
	binary.LittleEndian.PutUint32(b[nameHigh+4:nameHigh+8], uint32(query.NumberToReturn))

	queryHigh := nameHigh + 8 + len(query.query)
	copy(b[nameHigh+8:queryHigh], query.query)
	copy(b[queryHigh:], query.returnFieldsSelector)

	return b, nil
}

// Query returns the query document.
func (query *OpQuery) Query() *types.Document {
	if query.query == nil {
		return nil
	}

	return must.NotFail(query.query.Convert())
}

// ReturnFieldsSelector returns the fields selector document (that may be nil).
func (query *OpQuery) ReturnFieldsSelector() *types.Document {
	if query.returnFieldsSelector == nil {
		return nil
	}

	return must.NotFail(query.returnFieldsSelector.Convert())
}

// logMessage returns a string representation for logging.
func (query *OpQuery) logMessage(logFunc func(v any) string) string {
	if query == nil {
		return "<nil>"
	}

	m := must.NotFail(bson.NewDocument(
		"Flags", query.Flags.String(),
		"FullCollectionName", query.FullCollectionName,
		"NumberToSkip", query.NumberToSkip,
		"NumberToReturn", query.NumberToReturn,
	))

	doc, err := query.query.DecodeDeep()
	if err == nil {
		must.NoError(m.Add("Query", doc))
	} else {
		must.NoError(m.Add("QueryError", err.Error()))
	}

	if query.returnFieldsSelector != nil {
		doc, err = query.returnFieldsSelector.DecodeDeep()
		if err == nil {
			must.NoError(m.Add("ReturnFieldsSelector", doc))
		} else {
			must.NoError(m.Add("ReturnFieldsSelectorError", err.Error()))
		}
	}

	return logFunc(m)
}

// String returns a string representation for logging.
func (query *OpQuery) String() string {
	return query.logMessage(bson.LogMessage)
}

// StringBlock returns an indented string representation for logging.
func (query *OpQuery) StringBlock() string {
	return query.logMessage(bson.LogMessageBlock)
}

// StringFlow returns an unindented string representation for logging.
func (query *OpQuery) StringFlow() string {
	return query.logMessage(bson.LogMessageFlow)
}

// check interfaces
var (
	_ MsgBody = (*OpQuery)(nil)
)
