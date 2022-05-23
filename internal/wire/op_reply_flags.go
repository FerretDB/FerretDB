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

import "fmt"

//go:generate ../../bin/stringer -linecomment -type OpReplyFlagBit

// OpReplyFlagBit is a bit vector to specify OP_REPLY flags.
type OpReplyFlagBit flagBit

const (
	// OpReplyCursorNotFound indicates that the cursor is no longer exist at the server.
	OpReplyCursorNotFound = OpReplyFlagBit(1 << 0) // CursorNotFound

	// OpReplyQueryFailure reports query has failed.
	OpReplyQueryFailure = OpReplyFlagBit(1 << 1) // QueryFailure

	// OpReplyShardConfigStale indicates that shard config is outdated.
	OpReplyShardConfigStale = OpReplyFlagBit(1 << 2) // ShardConfigStale

	// OpReplyAwaitCapable indicates server supports AwaitData Query option.
	OpReplyAwaitCapable = OpReplyFlagBit(1 << 3) // AwaitCapable
)

// OpReplyFlags are OP_REPLY flags.
type OpReplyFlags flags

func opReplyFlagBitStringer(bit flagBit) string {
	return OpReplyFlagBit(bit).String()
}

// String returns string value for OP_REPLY.
func (f OpReplyFlags) String() string {
	return flags(f).string(opReplyFlagBitStringer)
}

// FlagSet returns true if the flag is set.
func (f OpReplyFlags) FlagSet(bit OpReplyFlagBit) bool {
	return f&OpReplyFlags(bit) != 0
}

// check interfaces
var (
	_ fmt.Stringer = OpReplyFlagBit(0)
	_ fmt.Stringer = OpReplyFlags(0)
)
