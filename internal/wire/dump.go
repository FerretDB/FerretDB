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
	"strings"

	"github.com/davecgh/go-spew/spew"
)

const useSpew = false

// DumpMsgHeader takes a MsgHeader and returns as a string.
func DumpMsgHeader(header *MsgHeader) string {
	var res string
	if useSpew {
		res = spew.Sdump(header)
	} else {
		b, err := json.MarshalIndent(header, "", "  ")
		if err != nil {
			panic(err)
		}
		res = string(b)
	}

	return strings.TrimSpace(res)
}

// DumpMsgBody takes a MsgBody and returns as a string.
func DumpMsgBody(body MsgBody) string {
	var res string
	if useSpew {
		res = spew.Sdump(body)
	} else {
		b, err := json.MarshalIndent(body, "", "  ")
		if err != nil {
			panic(err)
		}
		res = string(b)
	}

	return strings.TrimSpace(res)
}
