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

package common

import "fmt"

//go:generate ../../../bin/stringer -linecomment -type typeCode

// typeCode represents BSON type codes.
type typeCode int32

const (
	typeCodeDouble    = typeCode(1)  // double
	typeCodeString    = typeCode(2)  // string
	typeCodeObject    = typeCode(3)  // object
	typeCodeArray     = typeCode(4)  // array
	typeCodeBinData   = typeCode(5)  // binData
	typeCodeObjectID  = typeCode(7)  // objectId
	typeCodeBool      = typeCode(8)  // bool
	typeCodeDate      = typeCode(9)  // date
	typeCodeNull      = typeCode(10) // null
	typeCodeRegex     = typeCode(11) // regex
	typeCodeInt       = typeCode(16) // int
	typeCodeTimestamp = typeCode(17) // timestamp
	typeCodeLong      = typeCode(18) // long
	// not actual type code. number matches double, int and long
	typeCodeNumber = typeCode(-128) // number
)

// newTypeCode returns typeCde and error by given code.
func newTypeCode(code int32) (typeCode, error) {
	c := typeCode(code)
	switch c {
	case typeCodeDouble, typeCodeString, typeCodeObject, typeCodeArray,
		typeCodeBinData, typeCodeObjectID, typeCodeBool, typeCodeDate,
		typeCodeNull, typeCodeRegex, typeCodeInt, typeCodeTimestamp, typeCodeLong, typeCodeNumber:
		return c, nil
	case 19, -1, 127:
		return 0, NewErrorMsg(ErrNotImplemented, fmt.Sprintf(`Type code %v not implemented`, code))
	default:
		return 0, NewErrorMsg(ErrBadValue, fmt.Sprintf(`Invalid numerical type code: %d`, code))
	}
}

// aliasToTypeCode matches string type aliases to the corresponding typeCode value.
var aliasToTypeCode = map[string]typeCode{
	"double":    typeCodeDouble,
	"string":    typeCodeString,
	"object":    typeCodeObject,
	"array":     typeCodeArray,
	"binData":   typeCodeBinData,
	"objectId":  typeCodeObjectID,
	"bool":      typeCodeBool,
	"date":      typeCodeDate,
	"null":      typeCodeNull,
	"regex":     typeCodeRegex,
	"int":       typeCodeInt,
	"timestamp": typeCodeTimestamp,
	"long":      typeCodeLong,
	"number":    typeCodeNumber,
}
