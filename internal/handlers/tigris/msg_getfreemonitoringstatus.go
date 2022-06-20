package tigris

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/wire"
)

//TODO: add comments
func (h *Handler) MsgGetFreeMonitoringStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	return common.MsgGetFreeMonitoringStatus(ctx, msg)
}
