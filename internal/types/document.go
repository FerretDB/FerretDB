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

package types

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// isValidKey returns false if key is not a valid document field key.
func isValidKey(key string) bool {
	if key == "" {
		return false
	}

	return utf8.ValidString(key)
}

// Common interface with bson.Document.
type document interface {
	Map() map[string]any
	Keys() []string
}

// Document represents BSON document.
//
// Duplicate field names are not supported.
type Document struct {
	m    map[string]any
	keys []string
}

// ConvertDocument converts bson.Document to types.Document and validates it.
// It references the same data without copying it.
func ConvertDocument(d document) (Document, error) {
	if d == nil {
		panic("types.ConvertDocument: d is nil")
	}

	doc := Document{
		m:    d.Map(),
		keys: d.Keys(),
	}

	if doc.m == nil {
		doc.m = map[string]any{}
	}
	if doc.keys == nil {
		doc.keys = []string{}
	}

	if err := doc.validate(); err != nil {
		return doc, fmt.Errorf("types.ConvertDocument: %w", err)
	}

	return doc, nil
}

// MustConvertDocument is a ConvertDocument that panics in case of error.
func MustConvertDocument(d document) Document {
	doc, err := ConvertDocument(d)
	if err != nil {
		panic(err)
	}
	return doc
}

// MakeDocument makes a new Document from given key/value pairs.
func MakeDocument(pairs ...any) (Document, error) {
	l := len(pairs)
	if l%2 != 0 {
		return Document{}, fmt.Errorf("types.MakeDocument: invalid number of arguments: %d", l)
	}

	doc := Document{
		m:    make(map[string]any, l/2),
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

	if err := doc.validate(); err != nil {
		return doc, fmt.Errorf("types.MakeDocument: %w", err)
	}

	return doc, nil
}

// MustMakeDocument is a MakeDocument that panics in case of error.
func MustMakeDocument(pairs ...any) Document {
	doc, err := MakeDocument(pairs...)
	if err != nil {
		panic(err)
	}
	return doc
}

// validate checks if the document is valid.
func (d Document) validate() error {
	if len(d.m) != len(d.keys) {
		return fmt.Errorf("types.Document.validate: keys and values count mismatch: %d != %d", len(d.m), len(d.keys))
	}

	prevKeys := make(map[string]struct{}, len(d.keys))
	for _, key := range d.keys {
		if !isValidKey(key) {
			return fmt.Errorf("types.Document.validate: invalid key: %q", key)
		}

		value, ok := d.m[key]
		if !ok {
			return fmt.Errorf("types.Document.validate: key not found: %q", key)
		}

		if _, ok := prevKeys[key]; ok {
			return fmt.Errorf("types.Document.validate: duplicate key: %q", key)
		}
		prevKeys[key] = struct{}{}

		if err := validateValue(value); err != nil {
			return fmt.Errorf("types.Document.validate: %w", err)
		}
	}

	return nil
}

// Map returns a shallow copy of the document as a map. Do not modify it.
func (d Document) Map() map[string]any {
	return d.m
}

// Keys returns a shallow copy of the document's keys. Do not modify it.
func (d Document) Keys() []string {
	return d.keys
}

// Command returns the first document's key, this is often used as a command name.
func (d Document) Command() string {
	return strings.ToLower(d.keys[0])
}

func (d *Document) add(key string, value any) error {
	if _, ok := d.m[key]; ok {
		return fmt.Errorf("types.Document.add: key already present: %q", key)
	}

	if !isValidKey(key) {
		return fmt.Errorf("types.Document.add: invalid key: %q", key)
	}

	if err := validateValue(value); err != nil {
		return fmt.Errorf("types.Document.validate: %w", err)
	}

	d.keys = append(d.keys, key)
	d.m[key] = value

	return nil
}

// Get returns the value at the given key.
func (d Document) Get(key string) (any, error) {
	if value, ok := d.m[key]; ok {
		return value, nil
	}

	return nil, fmt.Errorf("types.Document.Get: key not found: %q", key)
}

// GetByPath returns Array or Object value by path - a sequence of indexes and keys.
func (d Document) GetByPath(path ...string) (any, error) {
	return getByPath(d, path...)
}

// Set the value of the given key, replacing any existing value.
func (d *Document) Set(key string, value any) error {
	if !isValidKey(key) {
		return fmt.Errorf("types.Document.Set: invalid key: %q", key)
	}

	if err := validateValue(value); err != nil {
		return fmt.Errorf("types.Document.validate: %w", err)
	}

	if _, ok := d.m[key]; !ok {
		d.keys = append(d.keys, key)
	}

	d.m[key] = value

	return nil
}

// Remove the given key, doing nothing if the key does not exist.
func (d *Document) Remove(key string) {
	if _, ok := d.m[key]; !ok {
		return
	}

	delete(d.m, key)

	for i, k := range d.keys {
		if k == key {
			d.keys = append(d.keys[:i], d.keys[i+1:]...)
			return
		}
	}

	// should not be reached
	panic(fmt.Sprintf("types.Document.Remove: key not found: %q", key))
}

// check interfaces
var (
	_ document = Document{}
	_ document = (*Document)(nil)
)
