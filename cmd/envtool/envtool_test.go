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
	"errors"
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

	err := rmdir([]string{"rz"})
	assert.NoError(t, err)
}

func TestMkdirAndRmdir(t *testing.T) {
	t.Parallel()

	paths := []string{"ab/c"}

	err := mkdir(paths)
	assert.NoError(t, err)

	// check if paths exist.
	var errs error

	for _, path := range paths {
		if _, err = os.Stat(path); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	assert.NoError(t, errs)

	err = rmdir(paths)
	assert.NoError(t, err)

	// check if paths do not exist.
	for _, path := range paths {
		if _, err = os.Stat(path); !os.IsNotExist(err) {
			errs = errors.Join(errs, err)
		}
	}

	assert.NoError(t, errs)
}
