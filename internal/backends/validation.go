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

package backends

import (
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

// databaseNameRe validates database name.
var databaseNameRe = regexp.MustCompile("^[a-zA-Z0-9_-]{1,63}$")

// collectionNameRe validates collection names.
var collectionNameRe = regexp.MustCompile("^[^\\.$\x00][^$\x00]{0,234}$")

// ReservedPrefix for names: databases, collections, schemas, tables, indexes, columns, etc.
const ReservedPrefix = "_ferretdb_"

// validateDatabaseName checks that database name is valid for FerretDB.
//
// It follows MongoDB restrictions plus
//   - allows only basic latin letters, digits, and basic punctuation;
//   - disallows `_ferretdb_` prefix.
//
// That validation is quite restrictive because
// we expect it to be easy for users to change database names in their software/configuration if needed.
//
// Backends can do their own additional validation.
func validateDatabaseName(name string) error {
	if !databaseNameRe.MatchString(name) {
		return NewError(ErrorCodeDatabaseNameIsInvalid, nil)
	}

	if strings.HasPrefix(name, ReservedPrefix) {
		return NewError(ErrorCodeDatabaseNameIsInvalid, nil)
	}

	return nil
}

// validateCollectionName checks that collection name is valid for FerretDB.
//
// It follows MongoDB restrictions plus:
//   - allows only UTF-8 characters;
//   - disallows '.' prefix (MongoDB fails to work with such collections correctly too);
//   - disallows `_ferretdb_` prefix.
//
// That validation is quite lax because
// we expect it to be hard for users to change collection names in their software.
//
// Backends can do their own additional validation.
func validateCollectionName(name string) error {
	if !collectionNameRe.MatchString(name) {
		return NewError(ErrorCodeCollectionNameIsInvalid, nil)
	}

	if strings.HasPrefix(name, ReservedPrefix) || strings.HasPrefix(name, "system.") {
		return NewError(ErrorCodeCollectionNameIsInvalid, nil)
	}

	if !utf8.ValidString(name) {
		return NewError(ErrorCodeCollectionNameIsInvalid, nil)
	}

	return nil
}

// validateSort validates if provided key and value are correct parts of the sort document.
// If they're not, it panics, thus the proper validation with error handling should be done
// properly on the handler level.
func validateSort(key string, value any) {
	sortValue := getWholeNumber(value)

	if sortValue != -1 && sortValue != 1 {
		panic("sort key ordering must be 1 (for ascending) or -1 (for descending)")
	}
}

// GetWholeNumber checks if the given value is int32, int64, or float64 containing a whole number,
// such as used in the sort, limit, $size, etc.
//
// It panics if provided value is invalid, thus the proper validation with error handling should be done
// properly on the handler level.
func getWholeNumber(value any) int64 {
	switch value := value.(type) {
	// add string support
	// TODO https://github.com/FerretDB/FerretDB/issues/1089
	case float64:
		switch {
		case math.IsInf(value, 1):
			panic("infinity")
		case value > float64(math.MaxInt64):
			panic("long exceeded - positive value")
		case value < float64(math.MinInt64):
			panic("long exceeded - negative value")
		case value != math.Trunc(value):
			panic("not a whole number")
		}

		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	default:
		panic("unexpected type")
	}
}
