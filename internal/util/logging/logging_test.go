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
	"bytes"
	"context"
	"encoding/hex"
	"log/slog"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapLogger(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, nil)).WithGroup("g1").With(slog.String("k1", "v1"))

	pc, _, _, _ := runtime.Caller(0)
	r := slog.NewRecord(time.Date(2024, 5, 31, 9, 26, 42, 0, time.UTC), slog.LevelInfo, "message", pc)

	wrappedL := WrapLogger(l)
	require.NoError(t, wrappedL.Handler().Handle(ctx, r))

	expected := "time=2024-05-31T09:26:42.000Z level=INFO msg=message g1.k1=v1\n"
	assert.Equal(t, expected, buf.String(), "actual:\n%s", hex.Dump(buf.Bytes()))
}
