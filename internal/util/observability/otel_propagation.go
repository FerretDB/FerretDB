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
	"encoding/hex"
	"encoding/json"

	"go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// TraceData represents OpenTelemetry tracing information that can be used to restore the info about parent span.
type TraceData struct {
	TraceID [16]byte `json:"traceID"`
	SpanID  [8]byte  `json:"spanID"`
}

// MarshalJSON implements json.Marshaler interface.
func (td *TraceData) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		TraceID string `json:"traceID"`
		SpanID  string `json:"spanID"`
	}{
		TraceID: hex.EncodeToString(td.TraceID[:]),
		SpanID:  hex.EncodeToString(td.SpanID[:]),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (td *TraceData) UnmarshalJSON(data []byte) error {
	var stringData struct {
		TraceID string `json:"traceID"`
		SpanID  string `json:"spanID"`
	}

	err := json.Unmarshal(data, &stringData)
	if err != nil {
		return lazyerrors.Error(err)
	}

	traceID, err := hex.DecodeString(stringData.TraceID)
	if err != nil {
		return lazyerrors.Error(err)
	}

	spanID, err := hex.DecodeString(stringData.SpanID)
	if err != nil {
		return lazyerrors.Error(err)
	}

	copy(td.TraceID[:], traceID)
	copy(td.SpanID[:], spanID)

	return nil
}

// SpanContextFromComment extracts OpenTelemetry tracing information from comment's field ferretDB.
// The comment is expected to be a string in JSON format.
//
// If the comment is empty or ferretDB field is not set, it returns an empty span context and no error.
func SpanContextFromComment(comment string) (trace.SpanContext, error) {
	if comment == "" {
		return trace.SpanContext{}, nil
	}

	type Comment struct {
		FerretDB *TraceData `json:"ferretDB"`
	}

	var data Comment

	err := json.Unmarshal([]byte(comment), &data)
	if err != nil {
		return trace.SpanContext{}, lazyerrors.Error(err)
	}

	if data.FerretDB == nil {
		return trace.SpanContext{}, nil
	}

	c := trace.SpanContextConfig{
		TraceID: trace.TraceID(data.FerretDB.TraceID),
		SpanID:  trace.SpanID(data.FerretDB.SpanID),
	}

	return trace.NewSpanContext(c), nil
}

// CommentFromSpanContext creates a comment string with OpenTelemetry tracing information set in comment's field ferretDB.
func CommentFromSpanContext(sc trace.SpanContext) string {
	if !sc.IsValid() {
		return ""
	}

	data := TraceData{
		TraceID: sc.TraceID(),
		SpanID:  sc.SpanID(),
	}

	comment, err := json.Marshal(&struct {
		FerretDB *TraceData `json:"ferretDB"`
	}{
		FerretDB: &data,
	})
	if err != nil {
		return ""
	}

	return string(comment)
}
