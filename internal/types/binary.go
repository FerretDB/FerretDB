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

type Binary struct {
	Subtype BinarySubtype
	B       []byte
}
