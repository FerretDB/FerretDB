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

package types

import (
	"bytes"
	"encoding/binary"
	"errors"
)

//go:generate ../../bin/stringer -linecomment -type BinarySubtype

// BinarySubtype represents BSON Binary's subtype.
type BinarySubtype byte

const (
	BinaryGeneric    = BinarySubtype(0x00) // generic
	BinaryFunction   = BinarySubtype(0x01) // function
	BinaryGenericOld = BinarySubtype(0x02) // generic-old
	BinaryUUIDOld    = BinarySubtype(0x03) // uuid-old
	BinaryUUID       = BinarySubtype(0x04) // uuid
	BinaryMD5        = BinarySubtype(0x05) // md5
	BinaryEncrypted  = BinarySubtype(0x06) // encrypted
	BinaryUser       = BinarySubtype(0x80) // user
)

// Binary represents BSON type Binary.
type Binary struct {
	Subtype BinarySubtype
	B       []byte
}

// BinaryFromArray takes position array which must contain non-negative numbers
// and packs it into types.Binary. Bit positions start at 0 from the least significant bit.
func BinaryFromArray(values *Array) (Binary, error) {
	var bitMask uint64
	for i := 0; i < values.Len(); i++ {
		value, err := values.Get(i)
		if err != nil {
			return Binary{}, err
		}

		if _, ok := value.(int32); !ok {
			return Binary{}, errors.New("bit position should be an integer value")
		}

		bitPosition := value.(int32)

		bitMask |= 1 << bitPosition
	}

	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, bitMask)

	return Binary{
		Subtype: BinaryGeneric,
		B:       bs,
	}, nil
}

// BinaryFromInt packs int64 value into types.Binary.
func BinaryFromInt(value int64) (Binary, error) {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.LittleEndian, uint64(value))
	if err != nil {
		return Binary{}, err
	}

	return Binary{
		Subtype: BinaryGeneric,
		B:       buff.Bytes(),
	}, nil
}
