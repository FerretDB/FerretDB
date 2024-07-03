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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestExporterWithFilter(t *testing.T) {
	t.Parallel()

	inMemExporter := tracetest.NewInMemoryExporter()

	exporter := ExporterWithFilter{exporter: inMemExporter}
	require.NotNil(t, exporter)

	root := tracetest.SpanStub{
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{SpanID: trace.SpanID{1}}),
	}
	span := tracetest.SpanStub{
		Parent:      root.Parent,
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{SpanID: trace.SpanID{11}}),
	}
	subspan := tracetest.SpanStub{
		Parent:      span.SpanContext,
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{SpanID: trace.SpanID{111}}),
	}
	excludedSpan := tracetest.SpanStub{
		Parent:      root.Parent,
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{SpanID: trace.SpanID{12}}),
		Attributes:  []attribute.KeyValue{ExclusionAttribute},
	}
	excludedSubspan := tracetest.SpanStub{
		Parent:      excludedSpan.SpanContext,
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{SpanID: trace.SpanID{121}}),
	}
	excludedSubspan2 := tracetest.SpanStub{
		Parent:      excludedSpan.SpanContext,
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{SpanID: trace.SpanID{122}}),
	}
	anotherSpan := tracetest.SpanStub{
		Parent:      root.Parent,
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{SpanID: trace.SpanID{13}}),
	}
	anotherSubspan := tracetest.SpanStub{
		Parent:      anotherSpan.SpanContext,
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{SpanID: trace.SpanID{131}}),
	}
	anotherExcludedSpan := tracetest.SpanStub{
		Parent:      root.Parent,
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{SpanID: trace.SpanID{14}}),
		Attributes:  []attribute.KeyValue{ExclusionAttribute},
	}

	spans := []tracetest.SpanStub{
		root, span, subspan, excludedSpan, excludedSubspan, excludedSubspan2,
		anotherSpan, anotherSubspan, anotherExcludedSpan,
	}

	require.NoError(t, exporter.ExportSpans(context.Background(), tracetest.SpanStubs(spans).Snapshots()))

	filteredSpans := inMemExporter.GetSpans()

	expectedSpanIDs := map[trace.SpanID]struct{}{
		root.SpanContext.SpanID():           {},
		span.SpanContext.SpanID():           {},
		subspan.SpanContext.SpanID():        {},
		anotherSpan.SpanContext.SpanID():    {},
		anotherSubspan.SpanContext.SpanID(): {},
	}

	require.Len(t, filteredSpans, len(expectedSpanIDs))

	for _, span := range filteredSpans {
		_, ok := expectedSpanIDs[span.SpanContext.SpanID()]
		assert.True(t, ok)
	}
}
