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

type attrs []groupOrAttrs

func (a attrs) toMap(r slog.Record) map[string]any {
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

func (a attrs) toDoc(r slog.Record) bson.D {
	m := a.toMap(r)
	var d bson.D

	for _, k := range slices.Sorted(maps.Keys(m)) {
		d = append(d, bson.E{k, resolveMapValue(m[k])})
	}

	return d
}

func resolveMapValue(v any) any {
	var m map[string]any
	var ok bool

	if m, ok = v.(map[string]any); !ok {
		return v
	}

	var d bson.D

	for _, k := range slices.Sorted(maps.Keys(m)) {
		d = append(d, bson.E{k, resolveMapValue(m[k])})
	}

	return d
}

//func (a attrs) toDoc(r slog.Record) bson.D {
//	elems := map[string]bson.E{}
//
//	r.Attrs(func(attr slog.Attr) bool {
//		if attr.Key != "" {
//			elems[attr.Key] = bson.E{attr.Key, resolveDoc(attr.Value)}
//			return true
//		}
//
//		if attr.Value.Kind() == slog.KindGroup {
//			for _, gAttr := range attr.Value.Group() {
//				elems[gAttr.Key] = bson.E{gAttr.Key, resolveDoc(gAttr.Value)}
//			}
//		}
//
//		return true
//	})
//
//	for _, goa := range slices.Backward(a) {
//		if goa.group != "" && len(elems) > 0 {
//			d := bson.D{}
//
//			sortedKeys := slices.Sorted(maps.Keys(elems))
//
//			for _, k := range sortedKeys {
//				d = append(d, elems[k])
//			}
//
//			elems = map[string]bson.E{goa.group: bson.E{goa.group, d}}
//			continue
//		}
//
//		for _, attr := range goa.attrs {
//			elems[attr.Key] = bson.E{attr.Key, resolveDoc(attr.Value)}
//		}
//	}
//
//	var d bson.D
//	for _, k := range slices.Sorted(maps.Keys(elems)) {
//		d = append(d, elems[k])
//	}
//
//	return d
//}

// attrs returns record attributes, as well as handler attributes from goas in map.
// Attributes with duplicate keys are overwritten, and the order of keys is ignored.
//
// TODO https://github.com/FerretDB/FerretDB/issues/4347
//func attrs(r slog.Record, goas []groupOrAttrs) map[string]any {
//	m := make(map[string]any, r.NumAttrs())
//
//	r.Attrs(func(attr slog.Attr) bool {
//		if attr.Key != "" {
//			m[attr.Key] = resolve(attr.Value)
//
//			return true
//		}
//
//		if attr.Value.Kind() == slog.KindGroup {
//			for _, gAttr := range attr.Value.Group() {
//				m[gAttr.Key] = resolve(gAttr.Value)
//			}
//		}
//
//		return true
//	})
//
//	for _, goa := range slices.Backward(goas) {
//		if goa.group != "" && len(m) > 0 {
//			m = map[string]any{goa.group: m}
//			continue
//		}
//
//		for _, attr := range goa.attrs {
//			m[attr.Key] = resolve(attr.Value)
//		}
//	}
//
//	return m
//}

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

// resolve returns underlying attribute value, or a map for [slog.KindGroup] type.
func resolveDoc(v slog.Value) any {
	v = v.Resolve()

	if v.Kind() != slog.KindGroup {
		return v.Any()
	}

	g := v.Group()
	d := make(bson.D, len(g))
	elems := map[string]bson.E{}

	for _, attr := range g {
		elems[attr.Key] = bson.E{attr.Key, resolveDoc(attr.Value)}
	}

	sortedKeys := slices.Sorted(maps.Keys(elems))
	for _, k := range sortedKeys {
		d = append(d, elems[k])
	}

	return d
}
