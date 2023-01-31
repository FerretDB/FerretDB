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
	"sort"
	"strconv"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Common interface with bson.Document.
//
// TODO Remove this type.
type document interface {
	Keys() []string
	Values() []any
}

// Document represents BSON document.
type Document struct {
	fields []field
}

// field represents a field in the document.
//
// The order of field is like that to reduce a pressure on gc a bit, and make vet/fieldalignment linter happy.
type field struct {
	value any
	key   string
}

// ConvertDocument converts bson.Document to *types.Document.
// It references the same data without copying it.
//
// TODO Remove this function: https://github.com/FerretDB/FerretDB/issues/260
func ConvertDocument(d document) (*Document, error) {
	if d == nil {
		panic("types.ConvertDocument: d is nil")
	}

	keys := d.Keys()
	values := d.Values()

	if len(keys) != len(values) {
		panic(fmt.Sprintf("document must have the same number of keys and values (keys: %d, values: %d)", len(keys), len(values)))
	}

	// If values are not set, we don't need to allocate memory for fields.
	if len(values) == 0 {
		return new(Document), nil
	}

	fields := make([]field, len(keys))
	for i, key := range d.Keys() {
		fields[i] = field{
			key:   key,
			value: values[i],
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

		doc.fields = append(doc.fields, field{key: key, value: value})
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

// Iterator returns an iterator over the document fields.
func (d *Document) Iterator() iterator.Interface[string, any] {
	return newDocumentIterator(d)
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
// If document or document's fields are not set (nil), it returns nil.
func (d *Document) Keys() []string {
	if d == nil || d.fields == nil {
		return nil
	}

	keys := make([]string, len(d.fields))
	for i, field := range d.fields {
		keys[i] = field.key
	}

	return keys
}

// Values returns a copy of document's values in the same order as Keys().
//
// If document or document's fields are not set (nil), it returns nil.
func (d *Document) Values() []any {
	if d == nil || d.fields == nil {
		return nil
	}

	values := make([]any, len(d.fields))
	for i, field := range d.fields {
		values[i] = field.value
	}

	return values
}

// FindDuplicateKey returns the first duplicate key in the document and true if duplicate exists.
// If duplicate keys don't exist it returns empty string and false.
func (d *Document) FindDuplicateKey() (string, bool) {
	seen := make(map[string]struct{}, len(d.fields))
	for _, field := range d.fields {
		if _, ok := seen[field.key]; ok {
			return field.key, true
		}

		seen[field.key] = struct{}{}
	}

	return "", false
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
// If the key is duplicated, it panics.
func (d *Document) Get(key string) (any, error) {
	if d.isKeyDuplicate(key) {
		panic(fmt.Sprintf("types.Document.Get: key is duplicated: %s", key))
	}

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

	for i, f := range d.fields {
		if f.key == key {
			d.fields[i].value = value
			return
		}
	}

	d.fields = append(d.fields, field{key: key, value: value})
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

		// In case if value is set in the middle of the array, we should fill the gap with Null
		for i := inner.Len(); i <= index; i++ {
			inner.Append(Null)
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

// SortFieldsByKey sorts the document fields by ascending order of the key.
func (d *Document) SortFieldsByKey() *Document {
	if d.Len() == 0 {
		return d
	}

	sorter := &fieldSorter{fields: d.fields}
	sorter.Sort()

	return d
}

// fieldSorter sorts the document fields by the key.
type fieldSorter struct {
	fields []field
}

// Sort the document field by key name.
func (ds *fieldSorter) Sort() {
	sort.Sort(ds)
}

// Len returns the length of the fields.
func (ds *fieldSorter) Len() int {
	return len(ds.fields)
}

// Swap changes position of elements i and j.
func (ds *fieldSorter) Swap(i, j int) {
	ds.fields[i], ds.fields[j] = ds.fields[j], ds.fields[i]
}

// Less returns true if the name of i-th key is less than j-th key.
func (ds *fieldSorter) Less(i, j int) bool {
	return Compare(ds.fields[i].key, ds.fields[j].key) == Less
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

// moveIDToTheFirstIndex sets the _id field of the document at the first position.
// If the _id field is not present, it does nothing.
func (d *Document) moveIDToTheFirstIndex() {
	if !d.Has("_id") {
		return
	}

	idIdx := 0

	if d.fields[idIdx].key == "_id" {
		return
	}

	for i, key := range d.Keys() {
		if key == "_id" {
			idIdx = i
			break
		}
	}

	d.fields = slices.Insert(d.fields, 0, field{key: d.fields[idIdx].key, value: d.fields[idIdx].value})

	d.fields = slices.Delete(d.fields, idIdx+1, idIdx+2)
}

// check interfaces
var (
	_ document = (*Document)(nil)
)
