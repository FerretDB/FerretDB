// Copyright 2021 Baltoro OÜ.
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

package types

import (
	"fmt"
	"strings"
)

// IsValidKey returns false if k is not a valid document field key.
func IsValidKey(key string) bool {
	// There are too many problems and edge cases with dots in field names;
	// disallow them for now.
	return key != "" && !strings.Contains(key, ".")
}

// Common interface with bson.Document.
type document interface {
	Map() map[string]interface{}
	Keys() []string
}

// Document represents BSON document.
//
// Duplicate field names are not supported.
type Document struct {
	m    map[string]interface{}
	keys []string
}

// NewDocuments makes a shallow copy of other Document or bson.Document.
func NewDocument(d document) Document {
	if d == nil {
		panic("d is nil")
	}

	res := Document{
		m:    d.Map(),
		keys: d.Keys(),
	}

	if res.m == nil {
		res.m = map[string]interface{}{}
	}
	if res.keys == nil {
		res.keys = []string{}
	}

	return res
}

// MakeDocument makes a new Document from given key/value pairs.
func MakeDocument(pairs ...interface{}) (Document, error) {
	l := len(pairs)
	if l%2 != 0 {
		return Document{}, fmt.Errorf("types.MakeDocument: invalid number of arguments: %d", l)
	}

	doc := Document{
		m:    make(map[string]interface{}, l/2),
		keys: make([]string, 0, l/2),
	}
	for i := 0; i < l; i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return Document{}, fmt.Errorf("types.MakeDocument: invalid key type: %T", pairs[i])
		}

		value := pairs[i+1]
		if err := doc.add(key, value); err != nil {
			return Document{}, fmt.Errorf("types.MakeDocument: %w", err)
		}
	}

	return doc, nil
}

// MustMakeDocument is a MakeDocument that panics in case of error.
func MustMakeDocument(pairs ...interface{}) Document {
	docs, err := MakeDocument(pairs...)
	if err != nil {
		panic(err)
	}
	return docs
}

// Map returns a shallow copy of the document as a map.
func (d Document) Map() map[string]interface{} {
	return d.m
}

// Keys returns a shallow copy of the document's keys.
func (d Document) Keys() []string {
	return d.keys
}

// Command returns the first documents's key that is often used as a command name.
func (d Document) Command() string {
	return strings.ToLower(d.keys[0])
}

func (d *Document) add(key string, value interface{}) error {
	if _, ok := d.m[key]; ok {
		return fmt.Errorf("Document.add: key already present: %q", key)
	}

	if !IsValidKey(key) {
		return fmt.Errorf("Document.add: invalid key: %q", key)
	}

	// TODO check value type

	d.keys = append(d.keys, key)
	d.m[key] = value

	return nil
}

// Set sets the value of the given key.
func (d *Document) Set(key string, value interface{}) error {
	if !IsValidKey(key) {
		return fmt.Errorf("Document.Set: invalid key: %q", key)
	}

	// TODO check value type

	if _, ok := d.m[key]; !ok {
		d.keys = append(d.keys, key)
	}

	d.m[key] = value

	return nil
}

// check interfaces
var (
	_ document = Document{}
	_ document = &Document{}
)
