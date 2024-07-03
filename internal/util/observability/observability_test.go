// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel/attribute"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
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
	exludedSubspan := tracetest.SpanStub{
		Parent:      excludedSpan.SpanContext,
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{SpanID: trace.SpanID{121}}),
	}

	spans := []tracetest.SpanStub{root, span, subspan, excludedSpan, exludedSubspan}

	require.NoError(t, exporter.ExportSpans(context.Background(), tracetest.SpanStubs(spans).Snapshots()))

	filteredSpans := inMemExporter.GetSpans()

	require.Len(t, filteredSpans, 3)
}
