package pg

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

func (h *Handler) MsgGetFreeMonitoringStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	var reply wire.OpMsg
	err := reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"argv", must.NotFail(types.NewArray("ferretdb")),
			"parsed", must.NotFail(types.NewDocument()),
			"ok", float64(1),
		))},
	})
	return
}
