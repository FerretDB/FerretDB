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

package lazyerrors

import "errors"

// UnwrapAll returns the last error in error chain, or nil, if err is nil.
func UnwrapAll(err error) error {
	if err == nil {
		return nil
	}

	for {
		e := errors.Unwrap(err)
		if e == nil {
			return err
		}
		err = e
	}
}
