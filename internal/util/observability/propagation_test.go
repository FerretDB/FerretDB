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

package observability

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestCommentFromSpanContext(t *testing.T) {
	traceID := [16]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef}
	spanID := [8]byte{0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10}

	sc := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	})

	comment, err := CommentFromSpanContext(sc)
	require.NoError(t, err)

	expectedComment := `{"ferretDB":{"traceID":"1234567890abcdef1234567890abcdef","spanID":"fedcba9876543210"}}`
	require.Equal(t, expectedComment, comment)

	parsed, err := SpanContextFromComment(comment)
	require.NoError(t, err)

	assert.Equal(t, sc.TraceID(), parsed.TraceID())
	assert.Equal(t, sc.SpanID(), parsed.SpanID())
}
