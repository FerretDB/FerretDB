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
	"fmt"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/users"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgHello implements `hello` command.
func (h *Handler) MsgHello(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	resp, err := h.hello(ctx, doc, h.TCPHost, h.ReplSetName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return wire.NewOpMsg(must.NotFail(resp.Encode()))
}

// hello checks client metadata and returns hello's document fields.
// It also returns response for deprecated `isMaster` and `ismaster` commands.
func (h *Handler) hello(ctx context.Context, spec bson.AnyDocument, tcpHost, name string) (*bson.Document, error) {
	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = checkClientMetadata(ctx, doc); err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := must.NotFail(bson.NewDocument())

	switch doc.Command() {
	case "hello":
		must.NoError(res.Add("isWritablePrimary", true))
	case "isMaster", "ismaster":
		if helloOk := doc.Get("helloOk"); helloOk != nil {
			must.NoError(res.Add("helloOk", true))
		}

		must.NoError(res.Add("ismaster", true))
	default:
		panic(fmt.Sprintf("unexpected command: %q", doc.Command()))
	}

	saslSupportedMechs, err := getOptionalParam(doc, "saslSupportedMechs", "")
	if err != nil {
		return nil, err
	}

	var resSupportedMechs *bson.Array

	if saslSupportedMechs != "" {
		db, username, ok := strings.Cut(saslSupportedMechs, ".")
		if !ok {
			return nil, handlererrors.NewCommandErrorMsg(
				handlererrors.ErrBadValue,
				"UserName must contain a '.' separated database.user pair",
			)
		}

		resSupportedMechs, err = h.getUserSupportedMechs(ctx, db, username)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	if name != "" {
		// That does not work for TLS-only setups, IPv6 addresses, etc.
		// The proper solution is to support `replSetInitiate` command.
		// TODO https://github.com/FerretDB/FerretDB/issues/3936
		if strings.HasPrefix(tcpHost, ":") {
			tcpHost = "localhost" + tcpHost
		}

		must.NoError(res.Add("setName", name))
		must.NoError(res.Add("hosts", must.NotFail(bson.NewArray(tcpHost))))
	}

	must.NoError(res.Add("maxBsonObjectSize", maxBsonObjectSize))
	must.NoError(res.Add("maxMessageSizeBytes", int32(wire.MaxMsgLen)))
	must.NoError(res.Add("maxWriteBatchSize", maxWriteBatchSize))
	must.NoError(res.Add("localTime", time.Now()))
	must.NoError(res.Add("logicalSessionTimeoutMinutes", logicalSessionTimeoutMinutes))
	must.NoError(res.Add("connectionId", connectionID))
	must.NoError(res.Add("minWireVersion", minWireVersion))
	must.NoError(res.Add("maxWireVersion", maxWireVersion))
	must.NoError(res.Add("readOnly", false))

	if resSupportedMechs != nil && resSupportedMechs.Len() != 0 {
		must.NoError(res.Add("saslSupportedMechs", resSupportedMechs))
	}

	must.NoError(res.Add("ok", float64(1)))

	return res, nil
}

// getUserSupportedMechs returns supported mechanisms for the given user.
// If the user was not found, it returns nil.
func (h *Handler) getUserSupportedMechs(ctx context.Context, dbName, username string) (*bson.Array, error) {
	filter, err := must.NotFail(bson.NewDocument("_id", dbName+"."+username)).Encode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	findSpec, err := must.NotFail(bson.NewDocument(
		"find", users.UserCollection,
		"filter", filter,
		"$db", users.UserDatabase,
	)).Encode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	page, err := h.Find(ctx, users.UserDatabase, findSpec)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := page.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	cursor, err := doc.Get("cursor").(bson.RawDocument).Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	firstBatchV := cursor.Get("firstBatch").(bson.RawArray)
	if firstBatchV == nil {
		return nil, nil
	}

	firstBatch, err := firstBatchV.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if firstBatch.Len() == 0 {
		return nil, nil
	}

	user, err := firstBatch.Get(0).(bson.RawDocument).Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	credentialsV := user.Get("credentials")
	if credentialsV == nil {
		return nil, nil
	}

	credentials := must.NotFail(credentialsV.(bson.RawDocument).Decode())

	supportedMechs := bson.MakeArray(len(credentials.FieldNames()))
	for _, mechanism := range credentials.FieldNames() {
		must.NoError(supportedMechs.Add(mechanism))
	}

	return supportedMechs, nil
}
