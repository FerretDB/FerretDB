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

package middleware

import (
	"context"
	"log/slog"
	"strconv"

	"go.opentelemetry.io/otel"
	otelattribute "go.opentelemetry.io/otel/attribute"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/observability"
)

// startSpan starts a new OpenTelemetry span for the request and returns the derived context.
func startSpan(ctx context.Context, req *Request, l *slog.Logger) context.Context {
	comment, _ := req.Document().Get("comment").(string)

	spanCtx, err := observability.SpanContextFromComment(comment)
	if err != nil {
		l.DebugContext(ctx, "Failed to extract span context from comment", logging.Error(err))
	}

	if spanCtx.IsValid() {
		ctx = oteltrace.ContextWithSpanContext(ctx, spanCtx)
	}

	command := req.Document().Command()
	database, _ := req.Document().Get("$db").(string)

	var collection string
	if command != "" {
		collection, _ = req.Document().Get(command).(string)
	}

	ctx, _ = otel.Tracer("").Start(
		ctx,
		command, // FIXME
		oteltrace.WithAttributes(
			otelsemconv.DBSystemNameKey.String("ferretdb"),
			otelsemconv.DBOperationName(command),
			otelsemconv.DBNamespace(database),
			otelsemconv.DBCollectionName(collection),
			otelattribute.Int("db.ferretdb.request_id", int(req.WireHeader().RequestID)),
		),
	)

	// Created span might be invalid, not sampled, and/or not recording,
	// if OpenTelemetry wasn't set up (for example, by the user of embeddable package).
	// We can't check span.SpanContext().IsValid(), span.SpanContext().IsSampled(), and span.IsRecording().

	return ctx
}

// endSpan ends the OpenTelemetry span from the context.
func endSpan(ctx context.Context, resp *Response) {
	span := oteltrace.SpanFromContext(ctx)

	if resp != nil {
		if c := resp.ErrorCode(); c != mongoerrors.ErrUnset {
			span.SetAttributes(
				otelsemconv.DBResponseStatusCode(strconv.Itoa(int(c))),
			)
		}

		span.SetAttributes(
			otelattribute.Int("db.ferretdb.response_id", int(resp.WireHeader().RequestID)),
		)
	}

	span.End()
}
