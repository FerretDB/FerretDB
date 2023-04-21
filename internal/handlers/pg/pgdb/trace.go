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

package pgdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("internal/handlers/pg/pgdb")

// debugTracer implements pgx.QueryTracer. It is used to add traces to Query,
// QueryRow, and Exec calls in debug builds.
type debugTracer struct{}

// TraceQueryStart adds a span to Query, QueryRow, and Exec calls.
func (t *debugTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx, _ = tracer.Start(ctx, data.SQL, trace.WithAttributes(
		attribute.String("args", fmt.Sprintf("%v", data.Args)),
	))

	return ctx
}

// TraceQueryEnd ends the span started by TraceQueryStart.
func (t *debugTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(attribute.String("commandTag", data.CommandTag.String()))

	if data.Err != nil {
		span.SetStatus(codes.Error, data.Err.Error())
	}

	span.End()
}

// multiQueryTracer implements pgx.QueryTracer. It can be used to add
// multiple tracers.
type multiQueryTracer struct {
	Tracers []pgx.QueryTracer
}

// TraceQueryStart starts all the tracers.
func (m *multiQueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	for _, t := range m.Tracers {
		ctx = t.TraceQueryStart(ctx, conn, data)
	}

	return ctx
}

// TraceQueryEnd ends all the tracers.
func (m *multiQueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	for _, t := range m.Tracers {
		t.TraceQueryEnd(ctx, conn, data)
	}
}
