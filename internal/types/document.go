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

// For compatibility with bson.Document.
type document interface {
	Map() map[string]interface{}
	Keys() []string
}

type Document struct {
	m    map[string]interface{}
	keys []string
}

func NewDocument(d document) Document {
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

func MakeDocument(pairs ...interface{}) Document {
	l := len(pairs)
	if l%2 != 0 {
		panic(fmt.Sprintf("invalid number of arguments: %d", l))
	}

	doc := Document{
		m:    make(map[string]interface{}, l/2),
		keys: make([]string, 0, l/2),
	}
	for i := 0; i < l; i += 2 {
		key := pairs[i].(string)
		value := pairs[i+1]
		doc.add(key, value)
	}

	return doc
}

func (d Document) Map() map[string]interface{} {
	return d.m
}

func (d Document) Keys() []string {
	return d.keys
}

func (d Document) Command() string {
	return strings.ToLower(d.keys[0])
}

func (d *Document) add(key string, value interface{}) {
	if _, ok := d.m[key]; ok {
		panic(fmt.Sprintf("key %q already present", key))
	}

	// TODO check value type

	d.keys = append(d.keys, key)
	d.m[key] = value
}

func (d *Document) Set(key string, value interface{}) {
	// TODO check value type

	if _, ok := d.m[key]; !ok {
		d.keys = append(d.keys, key)
	}

	d.m[key] = value
}

// check interfaces
var (
	_ document = Document{}
	_ document = &Document{}
)
