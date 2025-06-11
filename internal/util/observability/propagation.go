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

package observability

import (
	"encoding/json"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// commentData represents an operation comment formatted to contain tracing data.
type commentData struct {
	FerretDB struct {
		TraceID string `json:"traceID"`
		SpanID  string `json:"spanID"`
	} `json:"ferretDB"`
}

// SpanContextFromComment extracts OpenTelemetry tracing information from an operation comment.
// The comment is expected to be a json string (see commentData).
//
// If the comment is empty, it returns an empty span context and no error.
func SpanContextFromComment(comment string) (oteltrace.SpanContext, error) {
	var sc oteltrace.SpanContext

	if comment == "" {
		return sc, nil
	}

	var data commentData

	err := json.Unmarshal([]byte(comment), &data)
	if err != nil {
		return sc, lazyerrors.Error(err)
	}

	traceID, err := oteltrace.TraceIDFromHex(data.FerretDB.TraceID)
	if err != nil {
		return sc, lazyerrors.Error(err)
	}

	spanID, err := oteltrace.SpanIDFromHex(data.FerretDB.SpanID)
	if err != nil {
		return sc, lazyerrors.Error(err)
	}

	sc = oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	})

	return sc, nil
}

// CommentFromSpanContext creates a json-encoded string with tracing information (see commentData) from span context.
func CommentFromSpanContext(sc oteltrace.SpanContext) (string, error) {
	if !sc.IsValid() {
		return "", lazyerrors.New("invalid span context")
	}

	var data commentData
	data.FerretDB.TraceID = sc.TraceID().String()
	data.FerretDB.SpanID = sc.SpanID().String()

	comment, err := json.Marshal(data)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	return string(comment), nil
}
