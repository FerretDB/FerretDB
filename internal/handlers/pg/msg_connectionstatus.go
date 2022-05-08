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

package pg

import (
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgConnectionStatus
func (h *Handler) MsgConnectionStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	var reply wire.OpMsg

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	authInfo := must.NotFail(types.NewDocument(
		"authenticatedUsers", must.NotFail(types.NewArray()),
		"authenticatedUserRoles", must.NotFail(types.NewArray()),
	))

	showPrivileges, errMsg := getParamShowPrivileges(document)
	if errMsg != nil {
		err = reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"ok", float64(0),
			))},
		})
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		return &reply, errMsg
	}

	if showPrivileges {
		err := authInfo.Set("authenticatedUserPrivileges", must.NotFail(types.NewArray()))
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"authInfo", authInfo,
			"ok", float64(1),
		)),
		}})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// getParamShowPrivileges returns doc's value for key, default value for missing parameter, or protocol error for invalid parameter.
func getParamShowPrivileges(doc *types.Document) (bool, error) {
	v, err := doc.Get("showPrivileges")
	if err != nil {
		return false, nil
	}

	switch v := v.(type) {
	case float64:
		return v != float64(0), nil
	case bool:
		return v, nil
	case int32:
		return v != int32(0), nil
	case int64:
		return v != int64(0), nil
	case types.NullType:
		msg := fmt.Sprintf(`Expected boolean or number type for field "showPrivileges", found null`)
		return false, common.NewErrorMsg(common.ErrTypeMismatch, msg)
	default:
		msg := fmt.Sprintf("Expected boolean or number type for field \"showPrivileges\", found %T", v)
		return false, common.NewErrorMsg(common.ErrTypeMismatch, msg)
	}
}
