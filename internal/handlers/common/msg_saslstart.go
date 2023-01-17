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

package common

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgSASLStart is a common implementation of the SASLStart command.
func MsgSASLStart(_ context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	doc, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	_, err = GetRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	_, err = GetRequiredParam[int32](doc, "saslStart")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	_, err = GetRequiredParam[string](doc, "mechanism")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	_, err = GetRequiredParam[types.Binary](doc, "payload")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	_, err = GetOptionalParam(doc, "autoAuthorize", int32(0))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"done", true,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
