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
	"strconv"
	"time"

	"github.com/AlekSi/lazyerrors"
	"github.com/prometheus/client_golang/prometheus"
	otelattribute "go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// Dispatcher is a single-use object that sends a single request to a single handler,
// handling panics and response metrics, tracing, and logging.
// It is a part of the [Middleware], extracted to make it smaller.
type dispatcher struct {
	h         Handler
	l         *slog.Logger
	responses *prometheus.CounterVec
}

// Dispatch sends the request to the handler, handling panics and response metrics, tracing, and logging.
// It returns nil if unrecoverable error occurred.
func (d *dispatcher) Dispatch(ctx context.Context, req *Request) (resp *Response) {
	start := time.Now()

	var err error

	defer func() {
		var res result

		if p := recover(); p != nil {
			d.l.LogAttrs(ctx, logging.LevelDPanic, fmt.Sprintf("%[1]v (%[1]T)", p))

			res = resultPanic
			err = errors.New("panic")
		}

		resp, err = d.enforceContract(ctx, req, resp, err)

		switch err {
		case nil: // normal or error response
			if d.l.Enabled(ctx, slog.LevelDebug) {
				d.l.DebugContext(ctx, fmt.Sprintf(">>> %s\n%s", resp.WireHeader(), resp.WireBody().StringIndent()))
			}

			if resp.OK() {
				res = resultOK
				break
			}

			res = result(resp.ErrorName())

			if res == "" {
				d.l.LogAttrs(ctx, logging.LevelDPanic, "Unexpected result")
				res = resultUnknown
			}

		default: // unrecoverable error occurred
			if res == "" { // it might be set by panic
				res = resultError
			}
		}

		opcode := req.WireHeader().OpCode.String()
		command := req.Document().Command()

		var argument string
		if mErr := resp.MongoError(); mErr != nil {
			argument = mErr.Argument
		}

		if argument == "" {
			argument = "unknown"
		}

		// When both handlers are used, this metric is counted twice.
		// TODO https://github.com/FerretDB/FerretDB/issues/4987
		d.responses.With(prometheus.Labels{
			"opcode":   opcode,
			"command":  command,
			"argument": argument,
			"result":   string(res),
		}).Inc()

		d.endSpan(ctx, resp, res)

		attrs := []slog.Attr{
			slog.String("command", command),
			slog.String("result", string(res)),
			slog.String("duration", time.Since(start).String()),
		}
		if err != nil {
			attrs = append(attrs, logging.Error(err))
		}

		var level slog.Level

		switch res {
		case resultOK:
			level = slog.LevelInfo
		case resultError, resultPanic, resultUnknown:
			level = slog.LevelError
		default:
			level = slog.LevelWarn
		}

		d.l.LogAttrs(ctx, level, "Command handled", attrs...)
	}()

	resp, err = d.h.Handle(ctx, req)

	return
}

// enforceContract checks that [Handler]'s contract is not broken.
func (d *dispatcher) enforceContract(ctx context.Context, req *Request, resp *Response, err error) (*Response, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/4965
	level := logging.LevelDPanic

	if resp == nil && err == nil {
		msg := fmt.Sprintf("%T broke Handler contract: both are nil", d.h)
		d.l.LogAttrs(ctx, level, msg)

		return nil, lazyerrors.New(msg)
	}

	if resp != nil && err != nil {
		msg := fmt.Sprintf("%T broke Handler contract: both are non-nil", d.h)
		d.l.LogAttrs(ctx, level, msg, logging.Error(err))

		return nil, lazyerrors.New(msg)
	}

	var mongoErr *mongoerrors.Error
	if errors.As(err, &mongoErr) {
		msg := fmt.Sprintf("%T broke Handler contract: %T has %T in its chain", d.h, err, mongoErr)
		d.l.LogAttrs(ctx, level, msg, logging.Error(err))

		return ResponseErr(req, mongoErr), nil
	}

	return resp, err
}

// endSpan ends the OpenTelemetry span from the context (that is started in [Middleware.dispatch]).
func (d *dispatcher) endSpan(ctx context.Context, resp *Response, res result) {
	span := oteltrace.SpanFromContext(ctx)
	if res == resultOK {
		span.SetStatus(otelcodes.Ok, "")
	} else {
		span.SetStatus(otelcodes.Error, string(res))
	}

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
