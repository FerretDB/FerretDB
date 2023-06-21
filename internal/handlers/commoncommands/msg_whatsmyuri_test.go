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

package commoncommands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/wire"
)

// TestMsgWhatsMyURI checks a special case: if context is not set, it panics.
// The "normal" cases are covered in integration tests for MsgWhatsMyURI command.
func TestMsgWhatsMyURI(t *testing.T) {
	require.Panics(t, func() {
		_, err := MsgWhatsMyURI(context.Background(), new(wire.OpMsg))
		require.NoError(t, err)
	})
}
