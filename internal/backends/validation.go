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

// Reserved prefix for database and collection names.
const reservedPrefix = "_ferretdb_"

// databaseNameRe validates FerretDB database name.
var databaseNameRe = regexp.MustCompile("^[a-zA-Z0-9_-]{1,63}$")

// collectionNameRe validates collection names.
// Empty collection name, names with `$` and `\x00`,
// or exceeding the 255 bytes limit are not allowed.
// Collection names that start with `.` are also not allowed.
var collectionNameRe = regexp.MustCompile("^[^.$\x00][^$\x00]{0,234}$")

func validDatabaseName(name string) error {
	if databaseNameRe.MatchString(name) {
		return nil
	}

	return NewError(ErrorCodeDatabaseNameIsInvalid, nil)
}

func validCollectionName(name string) error {
	if !collectionNameRe.MatchString(name) {
		return NewError(ErrorCodeCollectionNameIsInvalid, nil)
	}
	if !utf8.ValidString(name) {
		return NewError(ErrorCodeCollectionNameIsInvalid, nil)
	}
	if strings.HasPrefix(name, reservedPrefix) {
		return NewError(ErrorCodeCollectionNameIsInvalid, nil)
	}

	return nil
}
