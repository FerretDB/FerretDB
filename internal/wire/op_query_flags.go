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

//go:generate ../../bin/stringer -linecomment -type OpQueryFlagBit

// OpQueryFlagBit an integer bitmask for the operation.
type OpQueryFlagBit flagBit

const (
	// OpQueryTailableCursor indicates that the cursor will not be closed.
	OpQueryTailableCursor = OpQueryFlagBit(1 << 1) // TailableCursor

	// OpQuerySlaveOk indicates that query can run on a replica slave.
	OpQuerySlaveOk = OpQueryFlagBit(1 << 2) // SlaveOk

	// OpQueryOplogReplay is deprecated.
	OpQueryOplogReplay = OpQueryFlagBit(1 << 3) // OplogReplay

	// OpQueryNoCursorTimeout disables cursor timeouts.
	OpQueryNoCursorTimeout = OpQueryFlagBit(1 << 4) // NoCursorTimeout

	// OpQueryAwaitData together with OpQueryTailableCursor, waits for data instead of returning it.
	OpQueryAwaitData = OpQueryFlagBit(1 << 5) // AwaitData

	// OpQueryExhaust indicates that server can divide data into multiple streams and expect that client can handle it.
	OpQueryExhaust = OpQueryFlagBit(1 << 6) // Exhaust

	// OpQueryPartial ignore error and give partial results.
	OpQueryPartial = OpQueryFlagBit(1 << 7) // Partial
)

// OpQueryFlags enables String() and FlagSet methods for flags.
type OpQueryFlags flags

func opQueryFlagBitStringer(bit flagBit) string {
	return OpQueryFlagBit(bit).String()
}

// String interface implementation for query flags.
func (f OpQueryFlags) String() string {
	return flags(f).string(opQueryFlagBitStringer)
}

// FlagSet return true if flag set.
func (f OpQueryFlags) FlagSet(bit OpQueryFlagBit) bool {
	return f&OpQueryFlags(bit) != 0
}

// check interfaces
var (
	_ fmt.Stringer = OpQueryFlagBit(0)
	_ fmt.Stringer = OpQueryFlags(0)
)
