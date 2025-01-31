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
	"sync"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mongoHandler struct {
	opts *NewHandlerOpts

	jsonHandler slog.Handler

	m   *sync.Mutex
	out io.Writer
}

type mongoLog struct {
	Timestamp  primitive.DateTime `bson:"t"`
	Severity   string             `bson:"s"`
	Components string             `bson:"c"`
	ID         int                `bson:"id"`
	Ctx        string             `bson:"ctx"`
	Svc        string             `bson:"svc"`
	Msg        string             `bson:"msg"`
	Attr       bson.D             `bson:"attr"`
	Tags       []string           `bson:"tags"`
	Truncated  bson.D             `bson:"truncated"`
	Size       bson.D             `bson:"size"`
}

func newMongoHandler(out io.Writer, opts *NewHandlerOpts) *mongoHandler {
	must.NotBeZero(opts)

	return &mongoHandler{
		opts: opts,
		m:    new(sync.Mutex),
		out:  out,
	}
}

func (h *mongoHandler) Enabled(_ context.Context, l slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}

	return l >= minLevel
}

func (h *mongoHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf bytes.Buffer

	logRecord := mongoLog{
		Timestamp: primitive.NewDateTimeFromTime(r.Time),
	}

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
	return h.jsonHandler.WithAttrs(attrs)
}

func (h *mongoHandler) WithGroup(name string) slog.Handler {
	return h.jsonHandler.WithGroup(name)
}
