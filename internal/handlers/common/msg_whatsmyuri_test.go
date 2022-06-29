package common

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
		_, err := MsgWhatsMyURI(context.Background(), &wire.OpMsg{})
		require.NoError(t, err)
	})
}
