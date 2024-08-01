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

package setup

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"log/slog"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/FerretDB/wire/wireclient"
	"github.com/stretchr/testify/require"
	"github.com/xdg-go/scram"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/password"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// setupClient returns test-specific non-authenticated low-level wire connection for the given MongoDB URI.
//
// It disconnects automatically when test ends.
//
// If the connection can't be established, it panics,
// as it doesn't make sense to proceed with other tests if we couldn't connect in one of them.
func setupWireConn(tb testtb.TB, ctx context.Context, uri string, l *slog.Logger) *wireclient.Conn {
	tb.Helper()

	ctx, span := otel.Tracer("").Start(ctx, "setupWireConn")
	defer span.End()

	conn, err := wireclient.Connect(ctx, uri, l)
	if err != nil {
		tb.Error(err)
		panic("setupWireConn: " + err.Error())
	}

	tb.Cleanup(func() {
		err = conn.Close()
		require.NoError(tb, err)
	})

	return conn
}

// authenticate verifies the provided credentials using the mechanism for the given connection.
func authenticate(ctx context.Context, c *wireclient.Conn, user string, pass password.Password, mech, authDB string) error {
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
