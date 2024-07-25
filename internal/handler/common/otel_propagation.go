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

package common

import (
	"encoding/json"

	"go.opentelemetry.io/otel/trace"
)

// SpanContextFromComment extracts OpenTelemetry tracing information from a comment document.
//
// If the comment is empty or parent field is not set, an empty span context is returned.
func SpanContextFromComment(comment string) (trace.SpanContext, error) {
	// TODO
	/*	if comment.Len() == 0 {
			return trace.SpanContext{}, nil
		}

		parent, err := GetRequiredParam[string](comment, "traceparent")
		if parent == "" || err != nil {
			return trace.SpanContext{}, nil
		}

		state, err := GetRequiredParam[string](comment, "tracestate")
		if err != nil {
			return trace.SpanContext{}, errors.New("missing tracestate")
		}*/

	type TraceData struct {
		TraceID [16]byte `json:"ferretTraceID"`
		SpanID  [8]byte  `json:"ferretSpanID"`
	}

	var data TraceData
	err := json.Unmarshal([]byte(comment), &data)
	if err != nil {
		return trace.SpanContext{}, nil
	}

	conf := trace.SpanContextConfig{
		TraceID: trace.TraceID(data.TraceID),
		SpanID:  trace.SpanID(data.SpanID),
	}

	return trace.NewSpanContext(conf), nil

	/*type TraceData struct {
		TraceParent string `json:"traceparent"`
		TraceState  string `json:"tracestate"`
	}

	var data TraceData
	err := json.Unmarshal([]byte(comment), &data)
	if err != nil {
		return trace.SpanContext{}, nil
	}

	// Fields are set according to https://opentelemetry.io/docs/specs/otel/context/api-propagators/#textmap-propagator.
	carrier := propagation.MapCarrier{
		"traceparent": data.TraceParent,
		"tracestate":  data.TraceState,
	}

	propagator := propagation.TraceContext{}
	ctx := context.Background()
	ctx = propagator.Extract(ctx, carrier)

	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return trace.SpanContext{}, errors.New("invalid span context")
	}

	return spanContext, nil*/
}
