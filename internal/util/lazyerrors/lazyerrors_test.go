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
	"github.com/stretchr/testify/require"
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

	require.Equal(t, "err", err.Error())
	require.Equal(t, "err1: err", err1.Error())
	require.Equal(t, "err2: err1: err", err2.Error())
	require.Equal(t, "err3: err2: err1: err", err3.Error())

	require.Equal(t, `&errors.errorString{s:"err"}`, fmt.Sprintf("%#v", err))
	require.Contains(t, fmt.Sprintf("%#v", err1), `&fmt.wrapError{msg:"err1: err", err:(*errors.errorString)(0x`)

	require.Equal(t, err, unwrap(err1, 1))
	require.Equal(t, nil, unwrap(err1, 2))

	require.Equal(t, err1, unwrap(err2, 1))
	require.Equal(t, err, unwrap(err2, 2))
	require.Equal(t, nil, unwrap(err2, 3))

	require.Equal(t, err2, unwrap(err3, 1))
	require.Equal(t, err1, unwrap(err3, 2))
	require.Equal(t, err, unwrap(err3, 3))
	require.Equal(t, nil, unwrap(err3, 4))

	require.True(t, errors.Is(err3, err3))
	require.True(t, errors.Is(err3, err2))
	require.True(t, errors.Is(err3, err1))
	require.True(t, errors.Is(err3, err))
}

func TestErrors(t *testing.T) {
	t.Parallel()

	err := New("err")
	err1 := Errorf("err1: %w", err)
	err2 := Errorf("err2: %w", err1)
	err3 := Errorf("err3: %w", err2)

	expected := "[lazyerrors_test.go:71 lazyerrors.TestErrors] err"
	require.Equal(t, expected, err.Error())
	expected = "[lazyerrors_test.go:72 lazyerrors.TestErrors] err1: " +
		"[lazyerrors_test.go:71 lazyerrors.TestErrors] err"
	require.Equal(t, expected, err1.Error())
	expected = "[lazyerrors_test.go:73 lazyerrors.TestErrors] err2: " +
		"[lazyerrors_test.go:72 lazyerrors.TestErrors] err1: " +
		"[lazyerrors_test.go:71 lazyerrors.TestErrors] err"
	require.Equal(t, expected, err2.Error())
	expected = "[lazyerrors_test.go:74 lazyerrors.TestErrors] err3: " +
		"[lazyerrors_test.go:73 lazyerrors.TestErrors] err2: " +
		"[lazyerrors_test.go:72 lazyerrors.TestErrors] err1: " +
		"[lazyerrors_test.go:71 lazyerrors.TestErrors] err"
	require.Equal(t, expected, err3.Error())

	expected = "lazyerror([lazyerrors_test.go:71 lazyerrors.TestErrors] err)"
	require.Equal(t, expected, fmt.Sprintf("%#v", err))
	expected = "lazyerror([lazyerrors_test.go:72 lazyerrors.TestErrors] err1: " +
		"[lazyerrors_test.go:71 lazyerrors.TestErrors] err)"
	require.Equal(t, expected, fmt.Sprintf("%#v", err1))
	expected = "lazyerror([lazyerrors_test.go:73 lazyerrors.TestErrors] err2: " +
		"[lazyerrors_test.go:72 lazyerrors.TestErrors] err1: " +
		"[lazyerrors_test.go:71 lazyerrors.TestErrors] err)"
	require.Equal(t, expected, fmt.Sprintf("%#v", err2))
	expected = "lazyerror([lazyerrors_test.go:74 lazyerrors.TestErrors] err3: " +
		"[lazyerrors_test.go:73 lazyerrors.TestErrors] err2: " +
		"[lazyerrors_test.go:72 lazyerrors.TestErrors] err1: " +
		"[lazyerrors_test.go:71 lazyerrors.TestErrors] err)"
	require.Equal(t, expected, fmt.Sprintf("%#v", err3))

	require.NotEqual(t, err, unwrap(err1, 1))
	require.Equal(t, err, unwrap(err1, 2))
	require.NotEqual(t, nil, unwrap(err1, 3))
	require.Equal(t, nil, unwrap(err1, 4))

	require.NotEqual(t, err1, unwrap(err2, 1))
	require.Equal(t, err1, unwrap(err2, 2))
	require.NotEqual(t, err, unwrap(err2, 3))
	require.Equal(t, err, unwrap(err2, 4))
	require.NotEqual(t, nil, unwrap(err2, 5))
	require.Equal(t, nil, unwrap(err2, 6))

	require.NotEqual(t, err2, unwrap(err3, 1))
	require.Equal(t, err2, unwrap(err3, 2))
	require.NotEqual(t, err1, unwrap(err3, 3))
	require.Equal(t, err1, unwrap(err3, 4))
	require.NotEqual(t, err, unwrap(err3, 5))
	require.Equal(t, err, unwrap(err3, 6))
	require.NotEqual(t, nil, unwrap(err3, 7))
	require.Equal(t, nil, unwrap(err3, 8))

	require.True(t, errors.Is(err3, err3))
	require.True(t, errors.Is(err3, err2))
	require.True(t, errors.Is(err3, err1))
	require.True(t, errors.Is(err3, err))
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
	assert.Equal(t, "[lazyerrors_test.go:142 lazyerrors.TestPC.func1] err", err.Error())
}

var drain any

func BenchmarkNew(b *testing.B) {
	for range b.N {
		drain = New("err")
	}

	b.StopTimer()

	assert.NotNil(b, drain)
}

func BenchmarkStatic(b *testing.B) {
	for range b.N {
		drain = errors.New("[lazyerrors_test.go:144 lazyerrors.BenchmarkStatic] err")
	}

	b.StopTimer()

	assert.NotNil(b, drain)
}
