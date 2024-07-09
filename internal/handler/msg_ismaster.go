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

package handler

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgIsMaster implements `isMaster` command.
func (h *Handler) MsgIsMaster(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	doc, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res, err := h.hello(ctx, doc, h.TCPHost, h.ReplSetName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.MakeOpMsgSection(res)))

	return &reply, nil
}

// checkClientMetadata checks if the message does not contain client metadata after it was received already.
func checkClientMetadata(ctx context.Context, doc *types.Document) error {
	c, _ := doc.Get("client")
	if c == nil {
		return nil
	}

	connInfo := conninfo.Get(ctx)
	if connInfo.MetadataRecv() {
		return handlererrors.NewCommandErrorMsg(
			handlererrors.ErrClientMetadataCannotBeMutated,
			"The client metadata document may only be sent in the first hello",
		)
	}

	connInfo.SetMetadataRecv()

	return nil
}
