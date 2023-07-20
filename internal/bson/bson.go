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

// Package bson provides converters from/to BSON for built-in and `types` types.
//
// See contributing guidelines and documentation for package `types` for details.
//
// # Mapping
//
// Composite types
//
//	Alias      types package    bson package
//
//	object     *types.Document  *bson.Document
//	array      *types.Array     *bson.arrayType
//
// Scalar types
//
//	Alias      types package    bson package
//
//	double     float64          *bson.doubleType
//	string     string           *bson.stringType
//	binData    types.Binary     *bson.binaryType
//	objectId   types.ObjectID   *bson.objectIDType
//	bool       bool             *bson.boolType
//	date       time.Time        *bson.dateTimeType
//	null       types.NullType   *bson.nullType
//	regex      types.Regex      *bson.regexType
//	int        int32            *bson.int32Type
//	timestamp  types.Timestamp  *bson.timestampType
//	long       int64            *bson.int64Type
//
//nolint:dupword // false positive
package bson

import (
	"bufio"
	"encoding"
	"fmt"
	"time"

	"github.com/AlekSi/pointer"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

type bsontype interface {
	bsontype() // seal for go-sumtype

	readNested(*bufio.Reader, int) error
	ReadFrom(*bufio.Reader) error
	WriteTo(*bufio.Writer) error
	encoding.BinaryMarshaler
}

//go-sumtype:decl bsontype

// TODO https://github.com/FerretDB/FerretDB/issues/260
func fromBSON(v bsontype) any {
	switch v := v.(type) {
	case *Document:
		return must.NotFail(types.ConvertDocument(v))
	case *arrayType:
		return pointer.To(types.Array(*v))
	case *doubleType:
		return float64(*v)
	case *stringType:
		return string(*v)
	case *binaryType:
		return types.Binary(*v)
	case *objectIDType:
		return types.ObjectID(*v)
	case *boolType:
		return bool(*v)
	case *dateTimeType:
		return time.Time(*v)
	case *nullType:
		return types.Null
	case *regexType:
		return types.Regex(*v)
	case *int32Type:
		return int32(*v)
	case *timestampType:
		return types.Timestamp(*v)
	case *int64Type:
		return int64(*v)
	case *CString:
		panic("CString should not be there")
	}

	panic(fmt.Sprintf("not reached: %T", v)) // for go-sumtype to work
}

// TODO https://github.com/FerretDB/FerretDB/issues/260
//
//nolint:deadcode,unused // remove later if it is not needed
func toBSON(v any) bsontype {
	switch v := v.(type) {
	case *types.Document:
		return MustConvertDocument(v)
	case *types.Array:
		return pointer.To(arrayType(*v))
	case float64:
		return pointer.To(doubleType(v))
	case string:
		return pointer.To(stringType(v))
	case types.Binary:
		return pointer.To(binaryType(v))
	case types.ObjectID:
		return pointer.To(objectIDType(v))
	case bool:
		return pointer.To(boolType(v))
	case time.Time:
		return pointer.To(dateTimeType(v))
	case types.NullType:
		return pointer.To(nullType(v))
	case types.Regex:
		return pointer.To(regexType(v))
	case int32:
		return pointer.To(int32Type(v))
	case types.Timestamp:
		return pointer.To(timestampType(v))
	case int64:
		return pointer.To(int64Type(v))
	}

	panic(fmt.Sprintf("not reached: %T", v)) // for go-sumtype to work
}
