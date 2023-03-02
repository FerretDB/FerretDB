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

package hana

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetFreeMonitoringStatus implements HandlerInterface.
func (h *Handler) MsgGetFreeMonitoringStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	// The state field in this function results in panic, though here it's used inside the stub,
	// which is not executed on runtime.
	return common.GetFreeMonitoringStatus(ctx, msg, nil)
}
