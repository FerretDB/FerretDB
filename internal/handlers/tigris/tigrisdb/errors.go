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
	"errors"
	"strings"

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

// isOtherCreationInFlight returns true if an attempt to create the database with the given name is already in progress.
// This function is implemented to keep nolint in a single place.
// TODO https://github.com/FerretDB/FerretDB/issues/1341
func isOtherCreationInFlight(err error) bool {
	var driverErr *driver.Error
	if !errors.As(err, &driverErr) {
		panic("isOtherCreationInFlight called with non-driver error")
	}

	isUnknnown := pointer.Get(driverErr).Code == api.Code_UNKNOWN //nolint:nosnakecase // Tigris named their const that way
	if !isUnknnown {
		return false
	}

	return strings.Contains(driverErr.Message, "duplicate key value, violates key constraint")
}

// IsInvalidArgument returns true if the error is "invalid argument" error.
// This function is implemented to keep nolint in a single place.
func IsInvalidArgument(err error) bool {
	e, _ := err.(*driver.Error)
	return pointer.Get(e).Code == api.Code_INVALID_ARGUMENT //nolint:nosnakecase // Tigris named their const that way
}
