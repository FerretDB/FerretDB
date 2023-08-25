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

package commonerrors

// writeError represents protocol write error.
// It required to build the correct write error result.
// The index field is optional and won't be used if it's nil.
type writeError struct {
	index  *int32
	errmsg string
	code   ErrorCode
}

// Error implements error interface.
func (we *writeError) Error() string {
	return we.errmsg
}

// check interfaces
var (
	_ error = (*writeError)(nil)
)
