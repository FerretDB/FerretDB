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
	"encoding/json"
	"log/slog"
	"slices"

	"go.mongodb.org/mongo-driver/bson"
)

// groupOrAttrs contains group name or attributes.
type groupOrAttrs struct {
	group string
	attrs []slog.Attr
}

type attributes []groupOrAttrs

func (*attributes) MarshalJSON() ([]byte, error) {
}

func (*attributes) MarshalBSON() ([]byte, error) {
}

var (
	_ json.Marshaler = (*attributes)(nil)
	_ bson.Marshaler = (*attributes)(nil)
)

func newAttrs(r slog.Record, goas []groupOrAttrs) attributes {
	a := goas[len(goas)-1]

	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key != "" {
			a.attrs = append(a.attrs, attr)
			return true
		}

		if attr.Value.Kind() == slog.KindGroup {
			for _, gAttr := range attr.Value.Group() {
				m[gAttr.Key] = resolve(gAttr.Value)
			}
		}

		return true
	})
}

// attrs returns record attributes, as well as handler attributes from goas in map.
// Attributes with duplicate keys are overwritten, and the order of keys is ignored.
//
// TODO https://github.com/FerretDB/FerretDB/issues/4347
func attrs(r slog.Record, goas []groupOrAttrs) map[string]any {
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

	for _, goa := range slices.Backward(goas) {
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
