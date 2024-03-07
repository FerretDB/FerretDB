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

package bson2

import (
	"bytes"
	"context"
	"log/slog"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

func unindent(t *testing.T, s string) string {
	t.Helper()

	parts := strings.Split(s, "\n")
	require.Positive(t, len(parts))
	if parts[0] == "" {
		parts = parts[1:]
	}

	indent := len(parts[0]) - len(strings.TrimLeft(parts[0], "\t"))
	require.GreaterOrEqual(t, indent, 0)

	for i := range parts {
		require.Greater(t, len(parts[i]), indent, "line: %q", parts[i])
		parts[i] = parts[i][indent:]
	}

	return strings.Join(parts, "\n")
}

func TestLogValue(t *testing.T) {
	opts := &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if groups != nil {
				return a
			}

			if a.Key == "v" {
				return a
			}

			return slog.Attr{}
		},
	}

	ctx := context.Background()

	var tbuf, jbuf bytes.Buffer
	tlog := slog.New(slog.NewTextHandler(&tbuf, opts))
	jlog := slog.New(slog.NewJSONHandler(&jbuf, opts))

	for _, tc := range []struct {
		name string
		v    slog.LogValuer
		t    string
		j    string
		m    string
	}{
		{
			name: "Numbers",
			v: must.NotFail(NewDocument(
				"f64", 42.0,
				"inf", float64(math.Inf(1)),
				"neg_inf", float64(math.Inf(-1)),
				"zero", math.Copysign(0, 1),
				"neg_zero", math.Copysign(0, -1),
				"nan", float64(math.NaN()),
				"i32", int32(42),
				"i64", int64(42),
			)),
			t: `v.f64=42 v.inf=+Inf v.neg_inf=-Inf v.zero=0 v.neg_zero=-0 v.nan=NaN v.i32=42 v.i64=42`,
			j: `{"v":{"f64":42,"inf":"+Inf","neg_inf":"-Inf","zero":0,"neg_zero":-0,"nan":"NaN","i32":42,"i64":42}}`,
			m: `
			{
			  "f64": 42.0,
			  "inf": +Inf,
			  "neg_inf": -Inf,
			  "zero": 0.0,
			  "neg_zero": -0.0,
			  "nan": NaN,
			  "i32": 42,
			  "i64": int64(42),
			}`,
		},
		{
			name: "Scalars",
			v: must.NotFail(NewDocument(
				"null", Null,
				"id", ObjectID{0x42},
				"bool", true,
				"time", time.Date(2023, 3, 6, 13, 14, 42, 123456789, time.FixedZone("", int(4*time.Hour.Seconds()))),
			)),
			t: `v.null=<nil> v.id=ObjectID(420000000000000000000000) v.bool=true v.time=2023-03-06T09:14:42.123Z`,
			j: `{"v":{"null":null,"id":"ObjectID(420000000000000000000000)","bool":true,"time":"2023-03-06T09:14:42.123Z"}}`,
			m: `
			{
			  "null": null,
			  "id": ObjectID(420000000000000000000000),
			  "bool": true,
			  "time": 2023-03-06T09:14:42.123Z,
			}`,
		},
		{
			name: "Composites",
			v: must.NotFail(NewDocument(
				"doc", must.NotFail(NewDocument(
					"foo", "bar",
					"baz", must.NotFail(NewDocument(
						"qux", "quux",
					)),
				)),
				"doc_raw", RawDocument{0x42},
				"doc_empty", must.NotFail(NewDocument()),
				"array", must.NotFail(NewArray(
					"foo",
					"bar",
					must.NotFail(NewArray("baz", "qux")),
				)),
			)),
			t: `v.doc.foo=bar v.doc.baz.qux=quux v.doc_raw=RawDocument<1> ` +
				`v.array.0=foo v.array.1=bar v.array.2.0=baz v.array.2.1=qux`,
			j: `{"v":{"doc":{"foo":"bar","baz":{"qux":"quux"}},"doc_raw":"RawDocument<1>",` +
				`"array":{"0":"foo","1":"bar","2":{"0":"baz","1":"qux"}}}}`,
			m: `
			{
			  "doc": {"foo": "bar", "baz": {"qux": "quux"}},
			  "doc_raw": RawDocument<1>,
			  "doc_empty": {},
			  "array": ["foo", "bar", ["baz", "qux"]],
			}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tlog.InfoContext(ctx, "", slog.Any("v", tc.v))
			assert.Equal(t, tc.t+"\n", tbuf.String())
			tbuf.Reset()

			jlog.InfoContext(ctx, "", slog.Any("v", tc.v))
			assert.Equal(t, tc.j+"\n", jbuf.String())
			jbuf.Reset()

			assert.Equal(t, unindent(t, tc.m), logMessage(tc.v))
		})
	}
}
