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
	"os"
	"testing"
	"testing/slogtest"
)

func TestEmbeddableHandler(t *testing.T) {
	t.Parallel()

	var testAttrs map[string]any

	newHandler := func(t *testing.T) slog.Handler {
		t.Helper()

		testAttrs = map[string]any{}

		opts := &NewHandlerOpts{
			Handler: slog.NewTextHandler(os.Stdout, new(slog.HandlerOptions)),
			Level:   slog.LevelDebug,
		}

		return newEmbeddableHandler(opts, testAttrs)
	}

	result := func(t *testing.T) map[string]any {
		t.Helper()

		return testAttrs
	}

	slogtest.Run(t, newHandler, result)
}
