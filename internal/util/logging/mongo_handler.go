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

package logging

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"maps"
	"runtime"
	"slices"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// mongoHandler is a [slog.Handler] that writes logs by using mongo structured JSON format.
// The format returns log entries with Relaxed Extended JSON specification.
// The format is not stable.
//
//nolint:vet // for readability
type mongoHandler struct {
	opts *NewHandlerOpts

	ga        []groupOrAttrs
	testAttrs map[string]any

	m   *sync.Mutex
	out io.Writer
}

// MongoLogRecord represents a single log message in mongo structured JSON format.
// For now it lacks the fields that are not used by any caller.
//
//nolint:vet // to preserve field ordering
type MongoLogRecord struct {
	Timestamp time.Time      `bson:"t"`
	Severity  string         `bson:"s"`
	Component string         `bson:"c"` // TODO https://github.com/FerretDB/FerretDB/issues/4431
	ID        int            `bson:"id,omitempty"`
	Ctx       string         `bson:"ctx"`
	Msg       string         `bson:"msg"`
	Attr      map[string]any `bson:"attr,omitempty"`
	Tags      []string       `bson:"tags,omitempty"`
}

// Marshal returns the mongo structured JSON encoding of log.
func (log *MongoLogRecord) Marshal() ([]byte, error) {
	return bson.MarshalExtJSON(&log, false, false)
}

// getSeverity maps logging levels to their mongo format counterparts.
// If provided level is not mapped, its standard format is returned.
func (log *MongoLogRecord) getSeverity(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "D"
	case slog.LevelInfo:
		return "I"
	case slog.LevelWarn:
		return "W"
	case slog.LevelError:
		return "E"
	default:
		return level.String()
	}
}

// mongoLogFromRecord constructs new MongoLogRecord based on provided slog.Record.
//
// When called by slog handler, handler's ga, and opts can be provided.
// If ga, or opts are nil, they're ignored.
func mongoLogFromRecord(r slog.Record, ga []groupOrAttrs, opts *NewHandlerOpts) *MongoLogRecord {
	if opts == nil {
		opts = new(NewHandlerOpts)
	}

	var log MongoLogRecord

	log.Msg = r.Message

	if !opts.RemoveLevel {
		log.Severity = log.getSeverity(r.Level)
	}

	if !opts.RemoveTime && !r.Time.IsZero() {
		log.Timestamp = r.Time
	}

	if !opts.RemoveSource && r.PC != 0 {
		f, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
		if f.File != "" {
			log.Ctx = shortPath(f.File) + ":" + strconv.Itoa(f.Line)
		}
	}

	log.Attr = attrs(r, ga)

	return &log
}

// newMongoHandler creates a new mongo handler.
func newMongoHandler(out io.Writer, opts *NewHandlerOpts, testAttrs map[string]any) *mongoHandler {
	must.NotBeZero(opts)

	return &mongoHandler{
		opts:      opts,
		testAttrs: testAttrs,
		m:         new(sync.Mutex),
		out:       out,
	}
}

// Enabled implements [slog.Handler].
func (h *mongoHandler) Enabled(_ context.Context, l slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}

	return l >= minLevel
}

// Handle implements [slog.Handler].
func (h *mongoHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf bytes.Buffer

	record := mongoLogFromRecord(r, h.ga, h.opts)

	if h.testAttrs != nil {
		if record.Msg != "" {
			h.testAttrs[slog.MessageKey] = record.Msg
		}

		if record.Severity != "" {
			h.testAttrs[slog.LevelKey] = record.Severity
		}

		if !record.Timestamp.IsZero() {
			h.testAttrs[slog.TimeKey] = record.Timestamp
		}

		if record.Ctx != "" {
			h.testAttrs[slog.SourceKey] = record.Ctx
		}

		if len(record.Attr) > 0 {
			maps.Copy(h.testAttrs, record.Attr)
		}
	}

	extJSON, err := record.Marshal()
	if err != nil {
		return err
	}

	_, err = buf.Write(extJSON)
	if err != nil {
		return err
	}

	buf.WriteRune('\n')

	h.m.Lock()
	defer h.m.Unlock()

	_, err = buf.WriteTo(h.out)

	return err
}

// WithAttrs implements [slog.Handler].
func (h *mongoHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	return &mongoHandler{
		opts:      h.opts,
		ga:        append(slices.Clone(h.ga), groupOrAttrs{attrs: attrs}),
		testAttrs: h.testAttrs,
		m:         h.m,
		out:       h.out,
	}
}

// WithGroup implements [slog.Handler].
func (h *mongoHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &mongoHandler{
		opts:      h.opts,
		ga:        append(slices.Clone(h.ga), groupOrAttrs{group: name}),
		testAttrs: h.testAttrs,
		m:         h.m,
		out:       h.out,
	}
}

// check interfaces
var (
	_ slog.Handler = (*mongoHandler)(nil)
)
