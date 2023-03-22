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
	"fmt"
	"strings"

	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
	"github.com/tigrisdata/tigris-client-go/driver"
)

// Make our own constants to avoid nosnakecase linter errors.
const (
	errUnknown         = api.Code_UNKNOWN
	errInvalidArgument = api.Code_INVALID_ARGUMENT
	errNotFound        = api.Code_NOT_FOUND
	errAlreadyExists   = api.Code_ALREADY_EXISTS
	errAborted         = api.Code_ABORTED
)

// IsInvalidArgument returns true if the error's code is errInvalidArgument.
func IsInvalidArgument(err error) bool {
	var driverErr *driver.Error
	if !errors.As(err, &driverErr) {
		panic(fmt.Sprintf("unexpected error type %#v", err))
	}

	return driverErr.Code == errInvalidArgument
}

// IsNotFound returns true if the error's code is errNotFound.
func IsNotFound(err error) bool {
	var driverErr *driver.Error
	if !errors.As(err, &driverErr) {
		return false
	}

	return driverErr.Code == errNotFound
}

// IsAlreadyExists returns true if the error's code is errAlreadyExists.
func IsAlreadyExists(err error) bool {
	var driverErr *driver.Error
	if !errors.As(err, &driverErr) {
		return false
	}

	return driverErr.Code == errAlreadyExists
}

// IsAborted returns true if the error's code is errAborted.
func IsAborted(err error) bool {
	var driverErr *driver.Error
	if !errors.As(err, &driverErr) {
		panic(fmt.Sprintf("unexpected error type %#v", err))
	}

	return driverErr.Code == errAborted
}

// isOtherCreationInFlight returns true if an attempt to create the database with the given name is already in progress.
func isOtherCreationInFlight(err error) bool {
	var driverErr *driver.Error
	if !errors.As(err, &driverErr) {
		panic(fmt.Sprintf("unexpected error type %#v", err))
	}

	if driverErr.Code != errUnknown {
		return false
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/1341
	return strings.Contains(driverErr.Message, "duplicate key value, violates key constraint")
}
