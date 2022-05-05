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

package dummy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
)

func TestDummyHandler(t *testing.T) {
	t.Parallel()

	h := New()
	ctx := context.Background()
	errNotImplemented := common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")
	for k, command := range common.Commands {
		if slices.Contains([]string{"debug_error", "debug_panic"}, k) {
			assert.NotNil(t, command.Handler)
			continue
		}
		if command.Handler != nil {
			_, err := command.Handler(h, ctx, nil)
			assert.Equal(t, err, errNotImplemented, k)
		}
	}
	_, err := h.CmdQuery(ctx, nil)
	assert.Equal(t, err, errNotImplemented)
}
