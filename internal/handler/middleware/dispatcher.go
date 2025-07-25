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

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/prometheus/client_golang/prometheus"
)

type dispatcher struct {
	h         Handler
	l         *slog.Logger
	responses *prometheus.CounterVec
}

var errPanic = errors.New("panic")

func (d *dispatcher) Handle(ctx context.Context, req *Request) (resp *Response, err error) {
	start := time.Now()

	defer func() {
		var result string

		if p := recover(); p != nil {
			d.l.LogAttrs(ctx, logging.LevelDPanic, fmt.Sprintf("%[1]v (%[1]T)", p))

			result = "panic"
			err = errPanic
		}

		var mongoErr *mongoerrors.Error
		if errors.As(err, &mongoErr) {
			msg := fmt.Sprintf("%T broke Handler contract: %T has %T in its chain", d.h, err, mongoErr)
			d.l.LogAttrs(ctx, logging.LevelDPanic, msg, logging.Error(err))
			resp = ResponseErr(req, mongoErr)
			err = nil
		}

		if err == nil {
			if resp.OK() {
				result = "ok"
			}
			if codeName, _ := resp.Document().Get("codeName").(string); codeName != "" {
				result = codeName
			}
		}

		if result == "" {
			result = "unknown"
		}

		// FIXME
		argument := "unknown"

		d.responses.With(prometheus.Labels{"argument": argument, "result": result}).Inc()

		// FIXME
		attrs := []slog.Attr{
			slog.String("command", req.doc.Command()),
			slog.String("result", result),
			slog.Duration("duration", time.Since(start)),
		}
		d.l.LogAttrs(ctx, slog.LevelInfo, "FIXME", attrs...)
	}()

	resp, err = d.h.Handle(ctx, req)
	return
}

// check interfaces
// FIXME
// var (
// 	_ Handler = (*dispatcher)(nil)
// )
