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

package pjson

import (
	"bytes"
	"encoding/json"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// schema describes document/object schema needed to unmarshal pjson document.
type schema struct {
	Properties map[string]*elem `json:"p"`  // document's properties
	Keys       []string         `json:"$k"` // to preserve properties' order
}

// elem describes an element of schema.
type elem struct {
	Type    elemType `json:"t"`            // for each field
	Schema  *schema  `json:"$s,omitempty"` // only for objects
	Options string   `json:"o,omitempty"`  // only for regex
	Items   []*elem  `json:"i,omitempty"`  // only for arrays
	Subtype byte     `json:"s,omitempty"`  // only for binData
}

// elemType represents possible types of schema elements.
type elemType string

// List of possible types in the schema elements.
const (
	elemTypeObject    elemType = "object"
	elemTypeArray     elemType = "array"
	elemTypeDouble    elemType = "double"
	elemTypeString    elemType = "string"
	elemTypeBinData   elemType = "binData"
	elemTypeObjectID  elemType = "objectId"
	elemTypeBool      elemType = "bool"
	elemTypeDate      elemType = "date"
	elemTypeNull      elemType = "null"
	elemTypeRegex     elemType = "regex"
	elemTypeInt       elemType = "int"
	elemTypeTimestamp elemType = "timestamp"
	elemTypeLong      elemType = "long"
)

// Schemas for scalar types.
var (
	doubleSchema = &elem{
		Type: elemTypeDouble,
	}
	stringSchema = &elem{
		Type: elemTypeString,
	}
	binDataSchema = func(subtype byte) *elem {
		return &elem{
			Type:    elemTypeBinData,
			Subtype: subtype,
		}
	}
	objectIDSchema = &elem{
		Type: elemTypeObjectID,
	}
	boolSchema = &elem{
		Type: elemTypeBool,
	}
	dateSchema = &elem{
		Type: elemTypeDate,
	}
	nullSchema = &elem{
		Type: elemTypeNull,
	}
	regexSchema = func(options string) *elem {
		return &elem{
			Type:    elemTypeRegex,
			Options: options,
		}
	}
	intSchema = &elem{
		Type: elemTypeInt,
	}
	timestampSchema = &elem{
		Type: elemTypeTimestamp,
	}
	longSchema = &elem{
		Type: elemTypeLong,
	}
)

// Marshal returns the JSON encoding of schema.
func (s *schema) Marshal() ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return b, nil
}

// Unmarshal parses the JSON-encoded schema.
func (s *schema) Unmarshal(b []byte) error {
	r := bytes.NewReader(b)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	if err := dec.Decode(s); err != nil {
		return lazyerrors.Error(err)
	}

	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
