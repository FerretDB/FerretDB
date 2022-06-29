package common

import (
	"context"
	"testing"

	"github.com/FerretDB/FerretDB/internal/wire"
	"github.com/stretchr/testify/require"
)

// TestMsgWhatsMyURI checks a special case: even if context is not set, the method shouldn't return an error or panic.
// The "normal" cases are covered in integration tests for MsgWhatsMyURI command.
func TestMsgWhatsMyURI(t *testing.T) {
	_, err := MsgWhatsMyURI(context.Background(), &wire.OpMsg{})
	require.NoError(t, err)
}
