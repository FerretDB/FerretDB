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

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestPrintDiagnosticData(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		l := testutil.Logger(t, zap.NewAtomicLevelAt(zap.DebugLevel))
		printDiagnosticData(nil, l.Sugar())
	})
}

func TestRmdirAbsentDir(t *testing.T) {
	t.Parallel()

	err := rmdir("absent")
	assert.NoError(t, err)
}

func TestMkdirAndRmdir(t *testing.T) {
	t.Parallel()

	paths := []string{"ab/c", "ab"}

	err := mkdir(paths...)
	assert.NoError(t, err)

	for _, path := range paths {
		assert.DirExists(t, path)
	}

	err = rmdir(paths...)
	assert.NoError(t, err)

	for _, path := range paths {
		assert.NoDirExists(t, path)
	}
}

func TestRead(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp("", "test_read")
	assert.NoError(t, err)

	s := "test string in a file"
	_, err = f.Write([]byte(s))
	assert.NoError(t, err)

	var output bytes.Buffer
	err = read(&output, f.Name())
	assert.NoError(t, err)
	assert.Equal(t, s, output.String())
}
