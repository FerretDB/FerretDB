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
	"encoding/json"
	"fmt"
)

//go:generate ../../bin/stringer -linecomment -type OpMsgFlagBit

// OpMsgFlagBit integer is a bitmask encoding flags that modify the format and behavior of OpMsg.
type OpMsgFlagBit flagBit

const (
	OpMsgChecksumPresent = OpMsgFlagBit(1 << 0)  // checksumPresent
	OpMsgMoreToCome      = OpMsgFlagBit(1 << 1)  // moreToCome
	OpMsgExhaustAllowed  = OpMsgFlagBit(1 << 16) // exhaustAllowed
)

func (i OpMsgFlagBit) MarshalJSON() ([]byte, error) {
	return []byte(`"` + i.String() + `"`), nil
}

// OpMsgFlags type unint32.
type OpMsgFlags flags

func opMsgFlagBitStringer(bit flagBit) string {
	return OpMsgFlagBit(bit).String()
}

// String returns OpMsgFlags as a string.
func (f OpMsgFlags) String() string {
	return flags(f).string(opMsgFlagBitStringer)
}

// MarshalJSON writes the OpMsgFlag in JSON format to a byte array.
func (f OpMsgFlags) MarshalJSON() ([]byte, error) {
	return json.Marshal(flags(f).strings(opMsgFlagBitStringer))
}

// FlagSet check if flag is set.
func (f OpMsgFlags) FlagSet(bit OpMsgFlagBit) bool {
	return f&OpMsgFlags(bit) != 0
}

// check interfaces
var (
	_ fmt.Stringer   = OpMsgFlagBit(0)
	_ json.Marshaler = OpMsgFlagBit(0)
	_ fmt.Stringer   = OpMsgFlags(0)
	_ json.Marshaler = OpMsgFlags(0)
)
