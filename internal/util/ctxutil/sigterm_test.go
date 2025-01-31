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

package ctxutil

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSigTerm(t *testing.T) {
	t.Run("Both", func(t *testing.T) {
		ctx1, stop1 := SigTerm(context.Background())
		defer stop1()

		ctx2, stop2 := SigTerm(context.Background())
		defer stop2()

		assert.Nil(t, ctx1.Err())
		assert.Nil(t, ctx2.Err())

		p, err := os.FindProcess(os.Getpid())
		require.NoError(t, err)

		err = p.Signal(os.Interrupt)
		require.NoError(t, err)

		<-ctx1.Done()
		<-ctx2.Done()

		assert.Equal(t, context.Canceled, ctx1.Err())
		assert.Equal(t, context.Canceled, ctx2.Err())
	})

	t.Run("Mix", func(t *testing.T) {
		ctx1, stop1 := SigTerm(context.Background())
		defer stop1()

		ctx2, stop2 := SigTerm(context.Background())
		defer stop2()

		assert.Nil(t, ctx1.Err())
		assert.Nil(t, ctx2.Err())

		stop1()

		assert.Equal(t, context.Canceled, ctx1.Err())
		assert.Nil(t, ctx2.Err())

		p, err := os.FindProcess(os.Getpid())
		require.NoError(t, err)

		err = p.Signal(os.Interrupt)
		require.NoError(t, err)

		<-ctx2.Done()

		assert.Equal(t, context.Canceled, ctx2.Err())
	})
}
