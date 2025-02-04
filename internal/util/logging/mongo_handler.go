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
	"runtime"
	"slices"
	"strconv"
	"sync"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mongoHandler struct {
	opts *NewHandlerOpts

	goas []groupOrAttrs
	m    *sync.Mutex
	out  io.Writer
}

type mongoLog struct {
	Timestamp  primitive.DateTime `bson:"t"`
	Severity   string             `bson:"s"`
	Components string             `bson:"c"`   //TODO
	ID         int                `bson:"id"`  //TODO
	Ctx        string             `bson:"ctx"` // TODO
	Svc        string             `bson:"svc,omitempty"`
	Msg        string             `bson:"msg"`
	Attr       map[string]any     `bson:"attr,omitempty"`
	Tags       []string           `bson:"tags,omitempty"`
	Truncated  bson.D             `bson:"truncated,omitempty"`
	Size       bson.D             `bson:"size,omitempty"`
}

func newMongoHandler(out io.Writer, opts *NewHandlerOpts) *mongoHandler {
	must.NotBeZero(opts)

	h := mongoHandler{
		opts: opts,
		m:    new(sync.Mutex),
		out:  out,
	}

	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}

	return &h
}

func (h *mongoHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.opts.Level.Level()
}

func (h *mongoHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf bytes.Buffer

	logRecord := mongoLog{
		Severity: getSeverity(r.Level),
		Msg:      r.Message,
	}

	if !r.Time.IsZero() {
		logRecord.Timestamp = primitive.NewDateTimeFromTime(r.Time)
	}

	if r.PC != 0 && !h.opts.RemoveSource {
		f, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
		if f.File != "" {
			logRecord.Ctx = shortPath(f.File) + ":" + strconv.Itoa(f.Line)
		}
	}

	m := make(map[string]any, r.NumAttrs())

	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key != "" {
			m[attr.Key] = resolve(attr.Value)

			return true
		}

		if attr.Value.Kind() == slog.KindGroup {
			for _, gAttr := range attr.Value.Group() {
				m[gAttr.Key] = resolve(gAttr.Value)
			}
		}

		return true
	})

	for _, goa := range slices.Backward(h.goas) {
		if goa.group != "" {
			m = map[string]any{goa.group: m}
			continue
		}

		for _, attr := range goa.attrs {
			m[attr.Key] = resolve(attr.Value)
		}
	}

	logRecord.Attr = m

	extJSON, err := bson.MarshalExtJSON(&logRecord, false, false)
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

func (h *mongoHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	return &mongoHandler{
		opts: h.opts,
		goas: append(slices.Clone(h.goas), groupOrAttrs{attrs: attrs}),
		m:    h.m,
		out:  h.out,
	}
}

func (h *mongoHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &mongoHandler{
		opts: h.opts,
		goas: append(slices.Clone(h.goas), groupOrAttrs{group: name}),
		m:    h.m,
		out:  h.out,
	}
}

func getSeverity(level slog.Level) string {
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
