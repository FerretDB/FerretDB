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
//
// Duplicate field names are not supported.
type Document struct {
	m    map[string]any
	keys []string
}

// ConvertDocument converts bson.Document to *types.Document.
// It references the same data without copying it.
//
// TODO Remove this function: https://github.com/FerretDB/FerretDB/issues/260
func ConvertDocument(d document) (*Document, error) {
	if d == nil {
		panic("types.ConvertDocument: d is nil")
	}

	doc := &Document{
		m:    d.Map(),
		keys: d.Keys(),
	}

	return doc, nil
}

// MakeDocument creates an empty document with set capacity.
func MakeDocument(capacity int) *Document {
	if capacity == 0 {
		return new(Document)
	}

	return &Document{
		m:    make(map[string]any, capacity),
		keys: make([]string, 0, capacity),
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
	return len(d.keys)
}

// Map returns this document as a map. Do not modify it.
//
// It returns nil for nil Document.
func (d *Document) Map() map[string]any {
	if d == nil {
		return nil
	}
	return d.m
}

// Keys returns document's keys. Do not modify it.
//
// It returns nil for nil Document.
func (d *Document) Keys() []string {
	if d == nil {
		return nil
	}
	return d.keys
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

// add adds the value for the given key, returning error if that key is already present.
//
// As a special case, _id always becomes the first key.
func (d *Document) add(key string, value any) error {
	if _, ok := d.m[key]; ok {
		return fmt.Errorf("types.Document.add: key already present: %q", key)
	}

	// update keys slice
	if key == "_id" {
		// TODO check that value is not regex or array: https://github.com/FerretDB/FerretDB/issues/1235

		// ensure that _id is the first field
		d.keys = slices.Insert(d.keys, 0, key)
	} else {
		d.keys = append(d.keys, key)
	}

	d.m[key] = value

	return nil
}

// Has returns true if the given key is present in the document.
func (d *Document) Has(key string) bool {
	_, ok := d.m[key]
	return ok
}

// Get returns a value at the given key.
func (d *Document) Get(key string) (any, error) {
	if value, ok := d.m[key]; ok {
		return value, nil
	}

	return nil, fmt.Errorf("types.Document.Get: key not found: %q", key)
}

// Set sets the value for the given key, replacing any existing value.
//
// As a special case, _id always becomes the first key.
func (d *Document) Set(key string, value any) error {
	// update keys slice
	if key == "_id" {
		// TODO check that value is not regex or array: https://github.com/FerretDB/FerretDB/issues/1235

		// ensure that _id is the first field
		if i := slices.Index(d.keys, key); i >= 0 {
			d.keys = slices.Delete(d.keys, i, i+1)
		}
		d.keys = slices.Insert(d.keys, 0, key)
	} else {
		if _, ok := d.m[key]; !ok {
			d.keys = append(d.keys, key)
		}
	}

	if d.m == nil {
		d.m = map[string]any{
			key: value,
		}
		return nil
	}

	d.m[key] = value
	return nil
}

// Remove the given key and return its value, or nil if the key does not exist.
func (d *Document) Remove(key string) any {
	if _, ok := d.m[key]; !ok {
		return nil
	}

	v := d.m[key]
	delete(d.m, key)

	for i, k := range d.keys {
		if k == key {
			d.keys = append(d.keys[:i], d.keys[i+1:]...)
			return v
		}
	}

	// should not be reached
	panic(fmt.Sprintf("types.Document.Remove: key not found: %q", key))
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
func (d *Document) SetByPath(path Path, value any) error {
	if path.Len() == 1 {
		return d.Set(path.Slice()[0], value)
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
		return inner.Set(path.Suffix(), value)
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

// check interfaces
var (
	_ document = (*Document)(nil)
)
