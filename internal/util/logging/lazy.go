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
	"log/slog"

	"github.com/FerretDB/wire/wirebson"
)

// lazyDecoder is a lazily evaluated [slog.LogValuer] for [wirebson.RawDocument]
// that tries to decode the document.
type lazyDecoder struct {
	raw  wirebson.RawDocument
	deep bool
}

// LogValue implements [slog.LogValuer].
func (ld lazyDecoder) LogValue() slog.Value {
	if len(ld.raw) == 0 {
		return slog.Value{}
	}

	var d *wirebson.Document
	if ld.deep {
		d, _ = ld.raw.DecodeDeep()
	} else {
		d, _ = ld.raw.Decode()
	}

	if d == nil {
		return ld.raw.LogValue()
	}

	return d.LogValue()
}

// LazyDecoder is a lazily evaluated [slog.LogValuer] for [wirebson.RawDocument]
// that tries to decode the document.
func LazyDecoder(raw wirebson.RawDocument) slog.LogValuer {
	return lazyDecoder{raw: raw, deep: false}
}

// LazyDeepDecoder is a lazily evaluated [slog.LogValuer] for [wirebson.RawDocument]
// that tries to deeply decode the document.
func LazyDeepDecoder(raw wirebson.RawDocument) slog.LogValuer {
	return lazyDecoder{raw: raw, deep: true}
}

// LazyString is a lazily evaluated [slog.LogValuer] for string.
type LazyString func() string

// LogValue implements [slog.LogValuer].
func (ls LazyString) LogValue() slog.Value { return slog.StringValue(ls()) }

// check interfaces
var (
	_ slog.LogValuer = lazyDecoder{}
	_ slog.LogValuer = (LazyString)(nil)
)
