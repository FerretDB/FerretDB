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
	"strconv"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Common interface with bson.Document.
//
// TODO Remove this type.
type document interface {
	Map() map[string]any
	Keys() []string
}

// Document represents BSON document.
type Document struct {
	fields []field
}

// field represents a field in the document.
type field struct {
	key   string
	value any
}

// ConvertDocument converts bson.Document to *types.Document.
// It references the same data without copying it.
//
// TODO Remove this function: https://github.com/FerretDB/FerretDB/issues/260
func ConvertDocument(d document) (*Document, error) {
	if d == nil {
		panic("types.ConvertDocument: d is nil")
	}

	// If both keys and map are nil, we don't need to allocate memory for fields.
	if d.Keys() == nil && d.Map() == nil {
		return new(Document), nil
	}

	m := d.Map()

	fields := make([]field, len(d.Keys()))
	for i, key := range d.Keys() {
		fields[i] = field{
			key:   key,
			value: m[key],
		}
	}

	return &Document{fields}, nil
}

// MakeDocument creates an empty document with set capacity.
func MakeDocument(capacity int) *Document {
	if capacity == 0 {
		return new(Document)
	}

	return &Document{
		fields: make([]field, 0, capacity),
	}
}

// NewDocument creates a document with the given key/value pairs.
func NewDocument(pairs ...any) (*Document, error) {
	l := len(pairs)
	if l%2 != 0 {
		return nil, fmt.Errorf("types.NewDocument: invalid number of arguments: %d", l)
	}

	doc := MakeDocument(l / 2)

	if l == 0 {
		return doc, nil
	}

	for i := 0; i < l; i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("types.NewDocument: invalid key type: %T", pairs[i])
		}

		value := pairs[i+1]
		if err := doc.add(key, value); err != nil {
			return nil, fmt.Errorf("types.NewDocument: %w", err)
		}
	}

	return doc, nil
}

func (*Document) compositeType() {}

// DeepCopy returns a deep copy of this Document.
func (d *Document) DeepCopy() *Document {
	if d == nil {
		panic("types.Document.DeepCopy: nil document")
	}
	return deepCopy(d).(*Document)
}

// Len returns the number of elements in the document.
//
// It returns 0 for nil Document.
func (d *Document) Len() int {
	if d == nil {
		return 0
	}

	return len(d.fields)
}

// Map returns this document as a map. Do not modify it.
//
// If there are duplicate keys in the document, the last value is set in the corresponding field.
//
// It returns nil for nil Document.
//
// Deprecated: as Document might have duplicate keys, map is not a good representation of it.
func (d *Document) Map() map[string]any {
	if d == nil {
		return nil
	}

	m := make(map[string]any, len(d.fields))
	for _, field := range d.fields {
		m[field.key] = field.value
	}

	return m
}

// Keys returns a copy of document's keys.
//
// If there are duplicate keys in the document, the result will have duplicate keys too.
//
// It returns nil for nil Document.
func (d *Document) Keys() []string {
	if d == nil {
		return nil
	}

	keys := make([]string, len(d.fields))
	for i, field := range d.fields {
		keys[i] = field.key
	}

	return keys
}

// Command returns the first document's key. This is often used as a command name.
// It returns an empty string if document is nil or empty.
func (d *Document) Command() string {
	keys := d.Keys()
	if len(keys) == 0 {
		return ""
	}
	return keys[0]
}

// add adds the value for the given key.
// If the key already exists, it will create a duplicate key.
//
// As a special case, _id always becomes the first key.
func (d *Document) add(key string, value any) error {
	if key == "_id" {
		// ensure that _id is the first field
		d.fields = slices.Insert(d.fields, 0, field{key, value})
	} else {
		d.fields = append(d.fields, field{key, value})
	}

	return nil
}

// Has returns true if the given key is present in the document.
func (d *Document) Has(key string) bool {
	for _, field := range d.fields {
		if field.key == key {
			return true
		}
	}

	return false
}

// Get returns a value at the given key.
// If there are duplicated keys in the document, it returns the first value.
func (d *Document) Get(key string) (any, error) {
	for _, field := range d.fields {
		if field.key == key {
			return field.value, nil
		}
	}

	return nil, fmt.Errorf("types.Document.Get: key not found: %q", key)
}

// Set sets the value for the given key, replacing any existing value.
// If the key is duplicated, it panics.
// If the key doesn't exist, it will be set.
//
// As a special case, _id always becomes the first key.
func (d *Document) Set(key string, value any) {
	if d.isKeyDuplicate(key) {
		panic(fmt.Sprintf("types.Document.Set: key is duplicated: %s", key))
	}

	if key == "_id" {
		// ensure that _id is the first field
		if i := slices.Index(d.Keys(), key); i >= 0 {
			d.fields = slices.Delete(d.fields, i, i+1)
		}
		d.fields = slices.Insert(d.fields, 0, field{key, value})

		return
	}

	for i, f := range d.fields {
		if f.key == key {
			d.fields[i].value = value
			return
		}
	}

	d.fields = append(d.fields, field{key, value})
}

// Remove the given key and return its value, or nil if the key does not exist.
// If the key is duplicated, it panics.
func (d *Document) Remove(key string) any {
	if d.isKeyDuplicate(key) {
		panic(fmt.Sprintf("types.Document.Remove: key is duplicated: %s", key))
	}

	for i, field := range d.fields {
		if field.key == key {
			d.fields = slices.Delete(d.fields, i, i+1)
			return field.value
		}
	}

	return nil
}

// HasByPath returns true if the given path is present in the document.
func (d *Document) HasByPath(path Path) bool {
	_, err := d.GetByPath(path)

	return err == nil
}

// GetByPath returns a value by path - a sequence of indexes and keys.
// If the Path has only one element, it returns the value for the given key.
func (d *Document) GetByPath(path Path) (any, error) {
	return getByPath(d, path)
}

// SetByPath sets value by given path. If the Path has only one element, it sets the value for the given key.
// If some parts of the path are missing, they will be created.
// The Document type will be used to create these parts.
// If multiple fields match the path it panics.
func (d *Document) SetByPath(path Path, value any) error {
	if path.Len() == 1 {
		d.Set(path.Slice()[0], value)
		return nil
	}

	if !d.HasByPath(path.TrimSuffix()) {
		// we should insert the missing part of the path
		if err := insertByPath(d, path); err != nil {
			return err
		}
	}

	innerComp := must.NotFail(d.GetByPath(path.TrimSuffix()))

	switch inner := innerComp.(type) {
	case *Document:
		inner.Set(path.Suffix(), value)
		return nil
	case *Array:
		index, err := strconv.Atoi(path.Suffix())
		if err != nil {
			return fmt.Errorf(
				"Cannot create field '%s' in element {%s: %s}",
				path.Suffix(),
				path.Slice()[len(path.Slice())-2],
				FormatAnyValue(innerComp),
			)
		}

		return inner.Set(index, value)
	default:
		return fmt.Errorf(
			"Cannot create field '%s' in element {%s: %s}",
			path.Suffix(),
			path.Prefix(),
			FormatAnyValue(must.NotFail(d.Get(path.Prefix()))),
		)
	}
}

// RemoveByPath removes document by path, doing nothing if the key does not exist.
// If the Path has only one element, it removes the value for the given key.
func (d *Document) RemoveByPath(path Path) {
	if path.Len() == 1 {
		d.Remove(path.Slice()[0])

		return
	}
	removeByPath(d, path)
}

// isKeyDuplicate returns true if the target key is duplicated in the document and false otherwise.
// If the key is not found, it returns false.
func (d *Document) isKeyDuplicate(targetKey string) bool {
	var found bool

	for _, key := range d.Keys() {
		if key == targetKey {
			if found {
				return true
			}

			found = true
		}
	}

	return false
}

// check interfaces
var (
	_ document = (*Document)(nil)
)
