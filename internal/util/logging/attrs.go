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

// Package logging provides logging helpers.
//
//nolint:forbidigo // bson.D needs to be used, as *wirebson.Document is not decodable by bson.Marshaler
package logging

import (
	"log/slog"
	"maps"
	"slices"

	"go.mongodb.org/mongo-driver/bson"
)

// groupOrAttrs contains group name or attributes.
type groupOrAttrs struct {
	group string
	attrs []slog.Attr
}

// attrsList contains a list of groupOrAttrs,
// ordered from the top level group to the latest one.
type attrsList []groupOrAttrs

// toMap returns record attributes, as well as handler attributes from attrList in map.
// Attributes with duplicate keys are overwritten.
func (a attrsList) toMap(r slog.Record) map[string]any {
	m := make(map[string]any, r.NumAttrs())

	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key != "" {
			m[attr.Key] = resolve(attr.Value)

			return true
		}

		if attr.Value.Kind() == slog.KindGroup {
			for _, gAttr := range attr.Value.Group() {
				m[gAttr.Key] = resolve(gAttr.Value)
			}
		}

		return true
	})

	for _, goa := range slices.Backward(a) {
		if goa.group != "" && len(m) > 0 {
			m = map[string]any{goa.group: m}
			continue
		}

		for _, attr := range goa.attrs {
			m[attr.Key] = resolve(attr.Value)
		}
	}

	return m
}

// toBSON returns record attributes, as well as handler attributes from attrList in bson.D.
// Attributes with duplicate keys are overwritten, and the elements are sorted by keys.
func (a attrsList) toBSON(r slog.Record) bson.D {
	docFields := map[string]bson.E{}

	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key != "" {
			docFields[attr.Key] = bson.E{Key: attr.Key, Value: resolveBSON(attr.Value)}
			return true
		}

		if attr.Value.Kind() == slog.KindGroup {
			for _, gAttr := range attr.Value.Group() {
				docFields[gAttr.Key] = bson.E{Key: gAttr.Key, Value: resolveBSON(gAttr.Value)}
			}
		}

		return true
	})

	for _, goa := range slices.Backward(a) {
		if goa.group != "" && len(docFields) > 0 {
			var groupDoc bson.D
			for _, k := range slices.Sorted(maps.Keys(docFields)) {
				groupDoc = append(groupDoc, docFields[k])
			}

			docFields = map[string]bson.E{goa.group: {Key: goa.group, Value: groupDoc}}

			continue
		}

		for _, attr := range goa.attrs {
			docFields[attr.Key] = bson.E{Key: attr.Key, Value: resolveBSON(attr.Value)}
		}
	}

	var outDoc bson.D
	for _, k := range slices.Sorted(maps.Keys(docFields)) {
		outDoc = append(outDoc, docFields[k])
	}

	return outDoc
}

// resolve returns underlying attribute value, or a map for [slog.KindGroup] type.
func resolve(v slog.Value) any {
	v = v.Resolve()

	if v.Kind() != slog.KindGroup {
		return v.Any()
	}

	g := v.Group()
	m := make(map[string]any, len(g))

	for _, attr := range g {
		m[attr.Key] = resolve(attr.Value)
	}

	return m
}

// resolveBSON returns underlying attribute value, or a sorted bson.D for [slog.KindGroup] type.
func resolveBSON(v slog.Value) any {
	v = v.Resolve()

	if v.Kind() != slog.KindGroup {
		return v.Any()
	}

	g := v.Group()

	var d bson.D
	elems := map[string]bson.E{}

	for _, attr := range g {
		elems[attr.Key] = bson.E{Key: attr.Key, Value: resolveBSON(attr.Value)}
	}

	for _, k := range slices.Sorted(maps.Keys(elems)) {
		d = append(d, elems[k])
	}

	return d
}
