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

package testutil

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/fuzz"
)

// WriteSeedCorpusFile adds given data to the fuzzing seed corpus for given fuzz function.
//
// It can be an alternative to using f.Add.
func WriteSeedCorpusFile(tb testing.TB, funcName string, b []byte) {
	tb.Helper()

	err := fuzz.Record(filepath.Join("testdata", "fuzz", funcName), b)
	require.NoError(tb, err)
}
