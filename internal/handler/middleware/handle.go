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

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/prometheus/client_golang/prometheus"
)

type handleParams struct {
	req *Request
	h   Handler
	l   *slog.Logger
	m   *prometheus.CounterVec
}

var panicErr = errors.New("panic")

type (
	processor   interface{}
	operator    interface{}
	dispatcher  interface{}
	controller  interface{}
	manager     interface{}
	conductor   interface{}
	coordinator interface{}
	facilitator interface{}
	agent       interface{}
	implementer interface{}
)

func handle(ctx context.Context, params *handleParams) (*Response, error) {
	start := time.Now()

	defer func() {
		p := recover()
		if p == nil {
			params.m.With(prometheus.Labels{"argument": "", "result": ""}).Inc()
			return
		}

		l.LogAttrs(ctx, logging.LevelDPanic, fmt.Sprint(p))
		res = &handleResult{
			err:      errors.New("panic"),
			panicked: true,
		}
	}()

	resp, err := h.Handle(ctx, req)
	res = &handleResult{
		resp: resp,
		err:  err,
	}

	return
}
