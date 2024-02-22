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

package lazyerrors

import (
	"errors"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func unwrap(err error, n int) error {
	for i := 0; i < n; i++ {
		err = errors.Unwrap(err)
	}
	return err
}

func TestStdErrors(t *testing.T) {
	t.Parallel()

	err := errors.New("err")
	err1 := fmt.Errorf("err1: %w", err)
	err2 := fmt.Errorf("err2: %w", err1)
	err3 := fmt.Errorf("err3: %w", err2)

	assert.Equal(t, "err", err.Error())
	assert.Equal(t, "err1: err", err1.Error())
	assert.Equal(t, "err2: err1: err", err2.Error())
	assert.Equal(t, "err3: err2: err1: err", err3.Error())

	assert.Equal(t, err, unwrap(err1, 1))
	assert.Equal(t, nil, unwrap(err1, 2))

	assert.Equal(t, err1, unwrap(err2, 1))
	assert.Equal(t, err, unwrap(err2, 2))
	assert.Equal(t, nil, unwrap(err2, 3))

	assert.Equal(t, err2, unwrap(err3, 1))
	assert.Equal(t, err1, unwrap(err3, 2))
	assert.Equal(t, err, unwrap(err3, 3))
	assert.Equal(t, nil, unwrap(err3, 4))

	assert.True(t, errors.Is(err3, err3))
	assert.True(t, errors.Is(err3, err2))
	assert.True(t, errors.Is(err3, err1))
	assert.True(t, errors.Is(err3, err))
}

func TestErrors(t *testing.T) {
	t.Parallel()

	err := New("err")
	err1 := Errorf("err1: %w", err)
	err2 := Errorf("err2: %w", err1)
	err3 := Errorf("err3: %w", err2)

	expected := "[lazyerrors_test.go:67 lazyerrors.TestErrors] err"
	assert.Equal(t, expected, err.Error())
	expected = "[lazyerrors_test.go:68 lazyerrors.TestErrors] err1: " +
		"[lazyerrors_test.go:67 lazyerrors.TestErrors] err"
	assert.Equal(t, expected, err1.Error())
	expected = "[lazyerrors_test.go:69 lazyerrors.TestErrors] err2: " +
		"[lazyerrors_test.go:68 lazyerrors.TestErrors] err1: " +
		"[lazyerrors_test.go:67 lazyerrors.TestErrors] err"
	assert.Equal(t, expected, err2.Error())
	expected = "[lazyerrors_test.go:70 lazyerrors.TestErrors] err3: " +
		"[lazyerrors_test.go:69 lazyerrors.TestErrors] err2: " +
		"[lazyerrors_test.go:68 lazyerrors.TestErrors] err1: " +
		"[lazyerrors_test.go:67 lazyerrors.TestErrors] err"
	assert.Equal(t, expected, err3.Error())

	assert.NotEqual(t, err, unwrap(err1, 1))
	assert.Equal(t, err, unwrap(err1, 2))
	assert.NotEqual(t, nil, unwrap(err1, 3))
	assert.Equal(t, nil, unwrap(err1, 4))

	assert.NotEqual(t, err1, unwrap(err2, 1))
	assert.Equal(t, err1, unwrap(err2, 2))
	assert.NotEqual(t, err, unwrap(err2, 3))
	assert.Equal(t, err, unwrap(err2, 4))
	assert.NotEqual(t, nil, unwrap(err2, 5))
	assert.Equal(t, nil, unwrap(err2, 6))

	assert.NotEqual(t, err2, unwrap(err3, 1))
	assert.Equal(t, err2, unwrap(err3, 2))
	assert.NotEqual(t, err1, unwrap(err3, 3))
	assert.Equal(t, err1, unwrap(err3, 4))
	assert.NotEqual(t, err, unwrap(err3, 5))
	assert.Equal(t, err, unwrap(err3, 6))
	assert.NotEqual(t, nil, unwrap(err3, 7))
	assert.Equal(t, nil, unwrap(err3, 8))

	assert.True(t, errors.Is(err3, err3))
	assert.True(t, errors.Is(err3, err2))
	assert.True(t, errors.Is(err3, err1))
	assert.True(t, errors.Is(err3, err))
}

func TestPC(t *testing.T) {
	t.Parallel()

	runtime.LockOSThread()

	ch := make(chan error, 1)

	go func() {
		runtime.LockOSThread()
		ch <- New("err")
	}()

	err := <-ch
	assert.Equal(t, "[lazyerrors_test.go:123 lazyerrors.TestPC.func1] err", err.Error())
}

var drain any

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		drain = New("err")
	}

	b.StopTimer()

	assert.NotNil(b, drain)
}

func BenchmarkStatic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		drain = errors.New("[lazyerrors_test.go:144 lazyerrors.BenchmarkStatic] err")
	}

	b.StopTimer()

	assert.NotNil(b, drain)
}
