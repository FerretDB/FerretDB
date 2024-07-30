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

// Package driver provides low-level wire protocol driver for testing.
package driver

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/FerretDB/wire/wireclient"
	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/password"
)

// Authenticate authenticates the given connection using the provided credentials and authentication mechanism.
func Authenticate(ctx context.Context, c *wireclient.Conn, user string, pass password.Password, mech, authDB string) error {
	password := pass.Password()

	var h scram.HashGeneratorFcn

	switch mech {
	case "SCRAM-SHA-1":
		h = scram.SHA1

		md5sum := md5.New()
		if _, err := md5sum.Write([]byte(user + ":mongo:" + password)); err != nil {
			return lazyerrors.Error(err)
		}

		src := md5sum.Sum(nil)
		dst := make([]byte, hex.EncodedLen(len(src)))
		hex.Encode(dst, src)

		password = string(dst)

	case "SCRAM-SHA-256":
		h = scram.SHA256

	default:
		return lazyerrors.Errorf("unsupported mechanism %q", mech)
	}

	s, err := h.NewClientUnprepped(user, password, "")
	if err != nil {
		return lazyerrors.Error(err)
	}

	conv := s.NewConversation()

	payload, err := conv.Step("")
	if err != nil {
		return lazyerrors.Error(err)
	}

	cmd := must.NotFail(wirebson.NewDocument(
		"saslStart", int32(1),
		"mechanism", mech,
		"payload", wirebson.Binary{B: []byte(payload)},
		"$db", authDB,
	))

	for {
		var body *wire.OpMsg

		if body, err = wire.NewOpMsg(must.NotFail(cmd.Encode())); err != nil {
			return lazyerrors.Error(err)
		}

		var resBody wire.MsgBody

		if _, resBody, err = c.Request(ctx, body); err != nil {
			return lazyerrors.Error(err)
		}

		var resMsg *wirebson.Document

		if resMsg, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode(); err != nil {
			return lazyerrors.Error(err)
		}

		if ok := resMsg.Get("ok"); ok != 1.0 {
			return lazyerrors.Errorf("%s was not successful (ok was %v)", cmd.Command(), ok)
		}

		if resMsg.Get("done").(bool) {
			return nil
		}

		payload, err = conv.Step(string(resMsg.Get("payload").(wirebson.Binary).B))
		if err != nil {
			return lazyerrors.Error(err)
		}

		cmd = must.NotFail(wirebson.NewDocument(
			"saslContinue", int32(1),
			"conversationId", int32(1),
			"payload", wirebson.Binary{B: []byte(payload)},
			"$db", authDB,
		))
	}
}
