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
	"regexp"
	"strings"
	"unicode/utf8"
)

// databaseNameRe validates database name.
var databaseNameRe = regexp.MustCompile("^[a-zA-Z0-9_-]{1,63}$")

// collectionNameRe validates collection names.
var collectionNameRe = regexp.MustCompile("^[^\\.$\x00][^$\x00]{0,234}$")

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

	if strings.HasPrefix(name, "_ferretdb_") {
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

	if strings.HasPrefix(name, "_ferretdb_") || strings.HasPrefix(name, "system.") {
		return NewError(ErrorCodeCollectionNameIsInvalid, nil)
	}

	if !utf8.ValidString(name) {
		return NewError(ErrorCodeCollectionNameIsInvalid, nil)
	}

	return nil
}

// validateIndexes checks that indexes are valid for FerretDB.
func validateIndexes(exitingIndexes, newIndexes []IndexInfo) error {
	for _, index := range newIndexes {
		if index.Name == "" {
			return NewError(ErrorCodeIndexNameIsEmpty, nil)
		}

		// Validation needs to be extended (e.g. to check that index names don't contain illegal symbols, see the issue).
		// TODO https://github.com/FerretDB/FerretDB/issues/3320

		for _, existing := range exitingIndexes {
			keyEqual := existing.Key.Equal(index.Key)

			if keyEqual && existing.Name == index.Name {
				if existing.Unique == index.Unique {
					// Indexes are equal, we don't need to create a new one, but we don't need to return an error.
					continue
				}

				return NewError(ErrorCodeIndexAlreadyExists, nil)
			}

			if keyEqual {
				return NewError(ErrorCodeIndexOptionsConflict, nil)
			}

			if existing.Name == index.Name {
				return NewError(ErrorCodeIndexKeySpecsConflict, nil)
			}
		}
	}

	return nil
}
