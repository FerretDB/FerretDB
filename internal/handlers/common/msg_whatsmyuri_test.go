package common

import (
	"context"
	"testing"

	"github.com/FerretDB/FerretDB/internal/wire"
)

func TestMsgWhatsMyURI(t *testing.T) {
	MsgWhatsMyURI(context.Background(), &wire.OpMsg{})
}
