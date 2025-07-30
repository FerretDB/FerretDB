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
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	otelattribute "go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/observability"
)

// Dispatcher is a single-use object that sends a single request to a single handler,
// handling panics, metrics, tracing, and logging.
// It is a part of the [Handler], extracted to make it smaller.
type dispatcher struct {
	h Handler
	l *slog.Logger
	m *Metrics
}

// errPanic is returned when a panic occurs in the handler.
var errPanic = errors.New("panic")

// Handle dispatches the request to the handler, handling panics, metrics, tracing, and logging.
func (d *dispatcher) Handle(ctx context.Context, req *Request) (resp *Response, err error) {
	start := time.Now()

	labels := prometheus.Labels{
		"opcode":  req.WireHeader().OpCode.String(),
		"command": req.Document().Command(),
	}
	d.m.requests.With(labels).Inc()

	ctx = d.startSpan(ctx, req)

	if d.l.Enabled(ctx, slog.LevelDebug) {
		d.l.DebugContext(ctx, fmt.Sprintf("<<< %s\n%s", req.WireHeader(), req.WireBody().StringIndent()))
	}

	defer func() {
		var res result

		if p := recover(); p != nil {
			d.l.LogAttrs(ctx, logging.LevelDPanic, fmt.Sprintf("%[1]v (%[1]T)", p))

			res = resultPanic
			err = errPanic
		}

		resp, err = d.enforceContract(ctx, req, resp, err)

		switch err {
		case nil:
			if resp.OK() {
				res = "ok"
				break
			}

			codeName, _ := resp.Document().Get("codeName").(string)
			if codeName != "" {
				res = result(codeName)
			}
		default:
			res = resultError
		}

		if res == "" {
			d.l.LogAttrs(ctx, logging.LevelDPanic, "Unexpected result")
			res = resultUnknown
		}

		// FIXME
		argument := "unknown"

		labels["argument"] = argument
		labels["result"] = string(res)
		d.m.responses.With(labels).Inc()

		d.endSpan(ctx, resp, res)

		if d.l.Enabled(ctx, slog.LevelDebug) {
			d.l.DebugContext(ctx, fmt.Sprintf(">>> %s\n%s", resp.WireHeader(), resp.WireBody().StringIndent()))
		}

		// FIXME
		attrs := []slog.Attr{
			slog.String("command", req.Document().Command()),
			slog.String("result", string(res)),
			slog.Duration("duration", time.Since(start)),
		}
		d.l.LogAttrs(ctx, slog.LevelInfo, "FIXME", attrs...)
	}()

	resp, err = d.h.Handle(ctx, req)
	return
}

// enforceContract checks that [Handler]'s contract is not broken.
func (d *dispatcher) enforceContract(ctx context.Context, req *Request, resp *Response, err error) (*Response, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/4965
	// level := logging.LevelDPanic
	level := logging.LevelWarn

	if resp == nil && err == nil {
		msg := fmt.Sprintf("%T broke Handler contract: both are nil", d.h)
		d.l.LogAttrs(ctx, level, msg)

		return nil, errPanic
	}

	if resp != nil && err != nil {
		msg := fmt.Sprintf("%T broke Handler contract: both are non-nil", d.h)
		d.l.LogAttrs(ctx, level, msg, logging.Error(err))

		return nil, errPanic
	}

	var mongoErr *mongoerrors.Error
	if errors.As(err, &mongoErr) {
		msg := fmt.Sprintf("%T broke Handler contract: %T has %T in its chain", d.h, err, mongoErr)
		d.l.LogAttrs(ctx, level, msg, logging.Error(err))

		return ResponseErr(req, mongoErr), nil
	}

	return resp, err
}

// startSpan starts a new OpenTelemetry span for the request and returns the derived context.
func (d *dispatcher) startSpan(ctx context.Context, req *Request) context.Context {
	comment, _ := req.Document().Get("comment").(string)

	spanCtx, err := observability.SpanContextFromComment(comment)
	if err != nil {
		d.l.DebugContext(ctx, "Failed to extract span context from comment", logging.Error(err))
	}
	if spanCtx.IsValid() {
		ctx = oteltrace.ContextWithSpanContext(ctx, spanCtx)
	}

	ctx, span := otel.Tracer("").Start(ctx, "")
	must.BeTrue(span.SpanContext().IsValid())
	must.BeTrue(span.SpanContext().IsRemote())
	must.BeTrue(span.SpanContext().IsSampled())
	must.BeTrue(span.IsRecording())

	span.SetName(req.doc.Command())

	span.SetAttributes(
		otelattribute.Int("db.ferretdb.request_id", int(req.header.RequestID)),
	)

	return ctx
}

// endSpan ends the OpenTelemetry span from the context.
func (d *dispatcher) endSpan(ctx context.Context, resp *Response, res result) {
	span := oteltrace.SpanFromContext(ctx)

	if res == resultOK {
		span.SetStatus(otelcodes.Ok, "")
	} else {
		span.SetStatus(otelcodes.Error, string(res))
	}

	span.SetAttributes(
		otelattribute.Int("db.ferretdb.response_id", int(resp.header.RequestID)),
	)

	span.End()
}
