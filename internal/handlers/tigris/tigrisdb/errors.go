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

package tigrisdb

import (
	"github.com/AlekSi/pointer"
	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
	"github.com/tigrisdata/tigris-client-go/driver"
)

// IsNotFound returns true if the error is "not found" error.
// This function is implemented to keep nolint in a single place.
func IsNotFound(err error) bool {
	e, _ := err.(*driver.Error)
	return pointer.Get(e).Code == api.Code_NOT_FOUND //nolint:nosnakecase // Tigris named their const that way
}

// IsAlreadyExists returns true if the error is "already exists" error.
// This function is implemented to keep nolint in a single place.
func IsAlreadyExists(err error) bool {
	e, _ := err.(*driver.Error)
	return pointer.Get(e).Code == api.Code_ALREADY_EXISTS //nolint:nosnakecase // Tigris named their const that way
}

// IsInvalidArgument returns true if the error is "invalid argument" error.
// This function is implemented to keep nolint in a single place.
func IsInvalidArgument(err error) bool {
	e, _ := err.(*driver.Error)
	return pointer.Get(e).Code == api.Code_INVALID_ARGUMENT //nolint:nosnakecase // Tigris named their const that way
}
