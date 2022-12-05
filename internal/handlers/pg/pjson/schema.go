package pjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// schema is a document's schema to unmarshal the document correctly.
type schema struct {
	Keys       []string         // $k
	Properties map[string]*elem // each elem from $k
}

// elem describes an element of schema.
type elem struct {
	Type       string    // type, for each field
	Properties *schema   // $s, only for objects
	Items      []*schema // $i, only for arrays
	Size       string    // s, only for binData
	Options    string    // o, only for regex
}

// marshalSchema marshals document's schema.
func marshalSchema(td *types.Document) (json.RawMessage, error) {
	var buf bytes.Buffer

	buf.WriteString(`{"$k":`)

	keys := td.Keys()
	if keys == nil {
		keys = []string{}
	}

	b, err := json.Marshal(keys)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	buf.Write(b)

	for _, key := range keys {
		buf.WriteByte(',')

		if b, err = json.Marshal(key); err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.Write(b)
		buf.WriteByte(':')

		value := must.NotFail(td.Get(key))

		switch val := value.(type) {
		case *types.Document:
			buf.WriteString(`{"t": "object", "$s":`)

			b, err := marshalSchema(val)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			buf.Write(b)

			buf.WriteByte('}')

		case *types.Array:
			buf.WriteString(`{"t": "array", "$i":`)

			// todo recursive schema for each element

			buf.WriteByte('}')

		case float64:
			buf.WriteString(`{"t": "double"}`)

		case string:
			buf.WriteString(`{"t": "string"}`)

		case types.Binary:
			buf.WriteString(`{"t": "binData", "s": 0}`) // todo

		case types.ObjectID:
			buf.WriteString(`{"t": "objectId"}`)

		case bool:
			buf.WriteString(`{"t": "bool"}`)

		case time.Time:
			buf.WriteString(`{"t": "date"}`)

		case types.NullType:
			buf.WriteString(`{"t": "null"}`)

		case types.Regex:
			buf.WriteString(`{"t": "regex", "o": ""}`) // todo

		case int32:
			buf.WriteString(`{"t": "int"}`)

		case types.Timestamp:
			buf.WriteString(`{"t": "timestamp"}`)

		case int64:
			buf.WriteString(`{"t": "long"}`)

		default:
			panic(fmt.Sprintf("pjson.marshalSchema: unknown type %[1]T (value %[1]q)", val))
		}

		b, err := Marshal(value)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.Write(b)
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

// unmarshalSchema unmarshals document's schema.
func unmarshalSchema(b json.RawMessage) {
	/*b, ok := b["$k"]
	if !ok {
		return lazyerrors.Errorf("pjson.documentType.UnmarshalJSON: missing $k")
	}*/
}
