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
	"errors"
	"sort"

	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListCommands returns a list of currently supported commands.
// Cannot add this func in Commands bcz of initialization loop.
func MsgListCommands(_ Handler, ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	var reply wire.OpMsg

	cmdList := must.NotFail(types.NewDocument())
	names := maps.Keys(Commands)
	sort.Strings(names)
	for _, name := range names {
		cmdList.Set(name, must.NotFail(types.NewDocument(
			"help", Commands[name].Help,
		)))
	}

	err := reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"commands", cmdList,
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// MsgDebugError used for debugging purposes.
func MsgDebugError(_ Handler, _ context.Context, _ *wire.OpMsg) (*wire.OpMsg, error) {
	return nil, errors.New("debug_error")
}

// MsgDebugPanic used for debugging purposes.
func MsgDebugPanic(_ Handler, _ context.Context, _ *wire.OpMsg) (*wire.OpMsg, error) {
	panic("debug_panic")
}
