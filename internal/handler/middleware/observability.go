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

	"github.com/FerretDB/wire"
	"go.opentelemetry.io/otel"
	otelattribute "go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/observability"
)

// MsgObservability is a middleware that will wrap the handler with logs, traces, and metrics.
//
// TODO https://github.com/FerretDB/FerretDB/issues/4439
func MsgObservability(next HandlerFunc[*MsgRequest, *MsgResponse], l *slog.Logger, command string) HandlerFunc[*MsgRequest, *MsgResponse] {
	return func(ctx context.Context, req *MsgRequest) (resp *MsgResponse, err error) {
		raw, _ := req.RawDocument()
		doc, _ := raw.Decode()
		comment, _ := doc.Get("comment").(string)

		if command == "" {
			command = "unknown"
		}

		var span oteltrace.Span
		ctx, span = startSpan(ctx, comment, l)

		defer func() {
			var result, argument string

			if resp.Error != nil {
				result = resp.Error.Name
				argument = resp.Error.Argument
			}

			if result == "" {
				result = "ok"
			}

			if argument == "" {
				argument = "unknown"
			}

			// TODO req.RequestID must be set on the request
			endSpan(span, command, wire.OpCodeMsg.String(), "ok", "unknown", req.RequestID)
		}()

		return next(ctx, req)
	}
}

// QueryObservability is a middleware that will wrap the handler with logs, traces, and metrics.
//
// TODO https://github.com/FerretDB/FerretDB/issues/4439
func QueryObservability(next HandlerFunc[*QueryRequest, *ReplyResponse], l *slog.Logger) HandlerFunc[*QueryRequest, *ReplyResponse] {
	return func(ctx context.Context, req *QueryRequest) (resp *ReplyResponse, err error) {
		command := "unknown"

		var span oteltrace.Span
		ctx, span = startSpan(ctx, "", l)

		defer func() {
			var result, argument string

			if resp.Error != nil {
				result = resp.Error.Name
				argument = resp.Error.Argument
			}

			if result == "" {
				result = "ok"
			}

			if argument == "" {
				argument = "unknown"
			}

			endSpan(span, command, wire.OpCodeReply.String(), result, argument, req.RequestID)
		}()

		return next(ctx, req)
	}
}

// startSpan gets the parent span from the comment field of the document,
// and starts a new span with it.
// If there is no span context, a new span without parent is started.
func startSpan(ctx context.Context, comment string, l *slog.Logger) (context.Context, oteltrace.Span) {
	var span oteltrace.Span

	spanCtx, err := observability.SpanContextFromComment(comment)
	if err != nil {
		l.DebugContext(ctx, "Failed to extract span context from comment", logging.Error(err))
		ctx, span = otel.Tracer("").Start(ctx, "")

		return ctx, span
	}

	ctx = oteltrace.ContextWithRemoteSpanContext(ctx, spanCtx)
	ctx, span = otel.Tracer("").Start(ctx, "")

	return ctx, span
}

// endSpan ends the span by setting status, name and attributes to the span.
func endSpan(span oteltrace.Span, command, resOpCode, result, argument string, responseTo int32) {
	must.NotBeZero(span)

	if result != "ok" {
		span.SetStatus(otelcodes.Error, result)
	}

	span.SetName(command)
	span.SetAttributes(
		otelattribute.String("db.ferretdb.opcode", resOpCode),
		otelattribute.Int("db.ferretdb.request_id", int(responseTo)),
		otelattribute.String("db.ferretdb.argument", argument),
	)
	span.End()
}
