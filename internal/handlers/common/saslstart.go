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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// SASLStart returns a common part of saslStart command response.
func SASLStart(ctx context.Context, doc *types.Document) error {
	mechanism, err := GetRequiredParam[string](doc, "mechanism")
	if err != nil {
		return lazyerrors.Error(err)
	}

	var username, password string

	switch mechanism {
	case "PLAIN":
		username, password, err = saslStartPlain(doc)
		if err != nil {
			return err
		}

	default:
		msg := fmt.Sprintf("Unsupported authentication mechanism %q.\n", mechanism) +
			"See https://docs.ferretdb.io/security/authentication/ for more details."
		return commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrAuthenticationFailed, msg, "mechanism")
	}

	conninfo.Get(ctx).SetAuth(username, password)

	return nil
}
