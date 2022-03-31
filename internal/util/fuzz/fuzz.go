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

package fuzz

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

func Record(dir string, b []byte) error {
	// https://github.com/golang/go/blob/a9d13a9c230bafba64469f126202315ba4d24eea/src/internal/fuzz/encoding.go
	var buf bytes.Buffer
	buf.WriteString("go test fuzz v1\n")
	if _, err := fmt.Fprintf(&buf, "[]byte(%q)\n", b); err != nil {
		return lazyerrors.Error(err)
	}

	if err := os.MkdirAll(dir, 0o777); err != nil {
		return lazyerrors.Error(err)
	}

	// https://github.com/golang/go/blob/378221bd6e73bdc21884fed9e32f53e6672ca0cd/src/internal/fuzz/fuzz.go
	filename := filepath.Join(dir, fmt.Sprintf("rec-%x", sha256.Sum256(b)))
	if err := os.WriteFile(filename, buf.Bytes(), 0o666); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
