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

package clientconn

import (
	"encoding/json"

	"go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// spanContextFromComment extracts OpenTelemetry tracing information from comment.
//
// If the comment is empty, it returns an empty span context and no error.
func spanContextFromComment(comment string) (trace.SpanContext, error) {
	if comment == "" {
		return trace.SpanContext{}, nil
	}

	type TraceData struct {
		TraceID [16]byte `json:"ferretTraceID"`
		SpanID  [8]byte  `json:"ferretSpanID"`
	}

	var data TraceData
	err := json.Unmarshal([]byte(comment), &data)
	if err != nil {
		return trace.SpanContext{}, lazyerrors.Error(err)
	}

	c := trace.SpanContextConfig{
		TraceID: trace.TraceID(data.TraceID),
		SpanID:  trace.SpanID(data.SpanID),
	}

	return trace.NewSpanContext(c), nil
}
