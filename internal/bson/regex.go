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

package bson

import (
	"bufio"
	"bytes"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// regexType represents BSON Regular expression type.
type regexType types.Regex

func (regex *regexType) bsontype() {}

// ReadFrom implements bsontype interface.
func (regex *regexType) ReadFrom(r *bufio.Reader) error {
	var pattern, options CString
	if err := pattern.ReadFrom(r); err != nil {
		return lazyerrors.Errorf("bson.Regex.ReadFrom (regex pattern): %w", err)
	}

	if err := options.ReadFrom(r); err != nil {
		return lazyerrors.Errorf("bson.Regex.ReadFrom (regex options): %w", err)
	}

	*regex = regexType{
		Pattern: string(pattern),
		Options: string(options),
	}
	return nil
}

// WriteTo implements bsontype interface.
func (regex regexType) WriteTo(w *bufio.Writer) error {
	v, err := regex.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.Regex.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.Regex.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (regex regexType) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	if err := CString(regex.Pattern).WriteTo(bufw); err != nil {
		return nil, err
	}
	if err := CString(regex.Options).WriteTo(bufw); err != nil {
		return nil, err
	}

	bufw.Flush()

	return buf.Bytes(), nil
}

// check interfaces
var (
	_ bsontype = (*regexType)(nil)
)
