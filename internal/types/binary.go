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

//go:generate ../../bin/stringer -linecomment -type BinarySubtype

// BinarySubtype represents BSON Binary's subtype.
type BinarySubtype byte

const (
	// BinaryGeneric represents a BSON generic binary subtype.
	BinaryGeneric = BinarySubtype(0x00) // generic

	// BinaryFunction represents a BSON function.
	BinaryFunction = BinarySubtype(0x01) // function

	// BinaryGenericOld represents a BSON generic-old.
	BinaryGenericOld = BinarySubtype(0x02) // generic-old

	// BinaryUUIDOld represents a BSON UUID old.
	BinaryUUIDOld = BinarySubtype(0x03) // uuid-old

	// BinaryUUID represents a BSON UUID.
	BinaryUUID = BinarySubtype(0x04) // uuid

	// BinaryMD5 represents a BSON md5.
	BinaryMD5 = BinarySubtype(0x05) // md5

	// BinaryEncrypted represents a Encrypted BSON value.
	BinaryEncrypted = BinarySubtype(0x06) // encrypted

	// BinaryUser represents a  User defined.
	BinaryUser = BinarySubtype(0x80) // user
)

// Binary represents BSON type Binary.
type Binary struct {
	B       []byte
	Subtype BinarySubtype
}
