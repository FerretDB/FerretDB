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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

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
	}{
		{
			name: "SimpleDoc",
			v:    must.NotFail(NewDocument("foo", "bar")),
			t:    "v.foo=bar",
			j:    `{"v":{"foo":"bar"}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tlog.InfoContext(ctx, "", slog.Any("v", tc.v))
			assert.Equal(t, tc.t+"\n", tbuf.String())
			tbuf.Reset()

			jlog.InfoContext(ctx, "", slog.Any("v", tc.v))
			assert.Equal(t, tc.j+"\n", jbuf.String())
			jbuf.Reset()
		})
	}
}
