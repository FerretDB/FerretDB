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

package tigris

import (
	"context"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDataSize implements HandlerInterface.
func (h *Handler) MsgDataSize(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := common.Unimplemented(document, "keyPattern", "min", "max"); err != nil {
		return nil, lazyerrors.Error(err)
	}
	common.Ignored(document, h.L, "estimate")

	m := document.Map()
	target, ok := m["dataSize"].(string)
	if !ok {
		return nil, lazyerrors.New("no target collection")
	}
	targets := strings.Split(target, ".")
	if len(targets) != 2 {
		return nil, lazyerrors.New("target collection must be like: 'database.collection'")
	}
	tdb, tcollection := targets[0], targets[1]

	// Count the time needed to fetch datasize.
	started := time.Now()

	// Retrieve the size in bytes using a dedicated tigris function.
	db := h.db.Driver.UseDatabase(tdb)
	collection, err := db.DescribeCollection(ctx, tcollection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	size := collection.Size

	// TODO We need a better way to get the number of documents in a collection.
	f := fetchParam{db: tdb, collection: tcollection}
	docs, err := h.fetch(ctx, f)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	objects := int32(len(docs))

	elapses := time.Since(started)

	var pairs []any
	if objects > 0 {
		pairs = append(pairs, "estimate", false)
	}
	pairs = append(pairs,
		"size", size,
		"numObjects", objects,
		"millis", int32(elapses.Milliseconds()),
		"ok", float64(1),
	)

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(pairs...))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
