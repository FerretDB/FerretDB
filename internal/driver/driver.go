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
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"path"
	"sync/atomic"

	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Conn represents a single connection.
//
// It is not safe for concurrent use.
//
// Logger is used only with debug level.
type Conn struct {
	c net.Conn
	r *bufio.Reader
	w *bufio.Writer
	l *slog.Logger

	authCreds     *url.Userinfo
	authMechanism string
	authDB        string
}

// Connect creates a new non-authenticated connection for the given MongoDB URI and logger.
//
// Context can be used to cancel the connection attempt.
// Canceling the context after the connection is established has no effect.
//
// Authentication credentials and mechanism can be set in the URI.
// They are not used by this function, but available to [Authenticate] and via [AuthInfo].
func Connect(ctx context.Context, uri string, l *slog.Logger) (*Conn, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if u.Scheme != "mongodb" {
		return nil, lazyerrors.Errorf("invalid scheme %q", u.Scheme)
	}

	if u.Opaque != "" {
		return nil, lazyerrors.Errorf("invalid URI %q", uri)
	}

	if _, _, err = net.SplitHostPort(u.Host); err != nil {
		return nil, lazyerrors.Error(err)
	}

	_, authDB := path.Split(u.Path)

	var authMechanism string

	for k, v := range u.Query() {
		switch k {
		case "authMechanism":
			authMechanism = v[0]

		case "authSource":
			authDB = v[0]

		case "replicaSet":
			// safe to ignore

		default:
			return nil, lazyerrors.Errorf("query parameter %q is not supported", k)
		}
	}

	l.DebugContext(ctx, "Connecting...", slog.String("uri", uri))

	d := net.Dialer{}

	c, err := d.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if authMechanism == "" {
		authMechanism = "SCRAM-SHA-1"
	}

	if authDB == "" {
		authDB = "admin"
	}

	return &Conn{
		c: c,
		r: bufio.NewReader(c),
		w: bufio.NewWriter(c),
		l: l,

		authCreds:     u.User,
		authMechanism: authMechanism,
		authDB:        authDB,
	}, nil
}

// AuthInfo returns the authentication credentials and mechanism extracted from the MongoDB URI.
func (c *Conn) AuthInfo() (*url.Userinfo, string) {
	return c.authCreds, c.authMechanism
}

// Authenticate attempts to authenticate the connection.
//
// This method is a shortcut for code that primary does not test authentication itself,
// but accesses MongoDB-compatible system that requires authentication.
// Code that tests various authentication scenarios should use [Request], [Write], or [WriteRaw] directly.
func (c *Conn) Authenticate(ctx context.Context) error {
	username := c.authCreds.Username()
	password, _ := c.authCreds.Password()

	var h scram.HashGeneratorFcn

	switch c.authMechanism {
	case "SCRAM-SHA-1":
		h = scram.SHA1

		md5sum := md5.New()
		if _, err := md5sum.Write([]byte(username + ":mongo:" + password)); err != nil {
			return lazyerrors.Error(err)
		}

		src := md5sum.Sum(nil)
		dst := make([]byte, hex.EncodedLen(len(src)))
		hex.Encode(dst, src)

		password = string(dst)

	case "SCRAM-SHA-256":
		h = scram.SHA256

	default:
		return lazyerrors.Errorf("unsupported mechanism %q", c.authMechanism)
	}

	s, err := h.NewClientUnprepped(username, password, "")
	if err != nil {
		return lazyerrors.Error(err)
	}

	conv := s.NewConversation()

	payload, err := conv.Step("")
	if err != nil {
		return lazyerrors.Error(err)
	}

	cmd := must.NotFail(bson.NewDocument(
		"saslStart", int32(1),
		"mechanism", c.authMechanism,
		"payload", bson.Binary{B: []byte(payload)},
		"$db", c.authDB,
	))

	for {
		var body *wire.OpMsg

		if body, err = wire.NewOpMsg(must.NotFail(cmd.Encode())); err != nil {
			return lazyerrors.Error(err)
		}

		var resBody wire.MsgBody

		if _, resBody, err = c.Request(ctx, nil, body); err != nil {
			return lazyerrors.Error(err)
		}

		var resMsg *bson.Document

		if resMsg, err = must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode(); err != nil {
			return lazyerrors.Error(err)
		}

		if ok := resMsg.Get("ok"); ok != 1.0 {
			return lazyerrors.Errorf("%s was not successful (ok was %v)", cmd.Command(), ok)
		}

		if resMsg.Get("done").(bool) {
			return nil
		}

		payload, err = conv.Step(string(resMsg.Get("payload").(bson.Binary).B))
		if err != nil {
			return lazyerrors.Error(err)
		}

		cmd = must.NotFail(bson.NewDocument(
			"saslContinue", int32(1),
			"conversationId", int32(1),
			"payload", bson.Binary{B: []byte(payload)},
			"$db", c.authDB,
		))
	}
}

// Close closes the connection.
func (c *Conn) Close() error {
	var err error

	c.l.Debug("Closing...")

	if e := c.w.Flush(); e != nil {
		err = lazyerrors.Error(e)
	}

	if e := c.c.Close(); e != nil && err == nil {
		err = lazyerrors.Error(e)
	}

	return err
}

// Read reads the next message from the connection.
func (c *Conn) Read() (*wire.MsgHeader, wire.MsgBody, error) {
	header, body, err := wire.ReadMessage(c.r)
	if err != nil {
		return nil, nil, lazyerrors.Error(err)
	}

	c.l.Debug(
		fmt.Sprintf("<<<\n%s", body.String()),
		slog.Int("length", int(header.MessageLength)),
		slog.Int("id", int(header.RequestID)),
		slog.Int("response_to", int(header.ResponseTo)),
		slog.String("opcode", header.OpCode.String()),
	)

	return header, body, nil
}

// Write writes the given message to the connection.
func (c *Conn) Write(header *wire.MsgHeader, body wire.MsgBody) error {
	c.l.Debug(
		fmt.Sprintf(">>>\n%s", body.String()),
		slog.Int("length", int(header.MessageLength)),
		slog.Int("id", int(header.RequestID)),
		slog.Int("response_to", int(header.ResponseTo)),
		slog.String("opcode", header.OpCode.String()),
	)

	if err := wire.WriteMessage(c.w, header, body); err != nil {
		return lazyerrors.Error(err)
	}

	if err := c.w.Flush(); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// WriteRaw writes the given raw bytes to the connection.
func (c *Conn) WriteRaw(b []byte) error {
	c.l.Debug(fmt.Sprintf(">>> %d raw bytes", len(b)))

	if _, err := c.w.Write(b); err != nil {
		return lazyerrors.Error(err)
	}

	if err := c.w.Flush(); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// lastRequestID stores incremented value of last recorded request header ID.
var lastRequestID atomic.Int32

// Request sends the given request to the connection and returns the response.
// If header MessageLength or RequestID is not specified, it assigns the proper values.
// For header.OpCode the wire.OpCodeMsg is used as default.
//
// It returns errors only for request/response parsing issues, or connection issues.
// All of the driver level errors are stored inside response.
func (c *Conn) Request(ctx context.Context, header *wire.MsgHeader, body wire.MsgBody) (*wire.MsgHeader, wire.MsgBody, error) {
	if header == nil {
		header = new(wire.MsgHeader)
	}

	if header.MessageLength == 0 {
		msgBin, err := body.MarshalBinary()
		if err != nil {
			return nil, nil, lazyerrors.Error(err)
		}

		header.MessageLength = int32(len(msgBin) + wire.MsgHeaderLen)
	}

	if header.OpCode == 0 {
		header.OpCode = wire.OpCodeMsg
	}

	if header.RequestID == 0 {
		header.RequestID = lastRequestID.Add(1)
	}

	if header.ResponseTo != 0 {
		return nil, nil, lazyerrors.Errorf("setting response_to is not allowed")
	}

	if m, ok := body.(*wire.OpMsg); ok {
		if m.Flags != 0 {
			return nil, nil, lazyerrors.Errorf("unsupported request flags %s", m.Flags)
		}
	}

	if err := c.Write(header, body); err != nil {
		return nil, nil, lazyerrors.Error(err)
	}

	resHeader, resBody, err := c.Read()
	if err != nil {
		return nil, nil, lazyerrors.Error(err)
	}

	if resHeader.ResponseTo != header.RequestID {
		c.l.ErrorContext(
			ctx,
			"response_to is not equal to request_id",
			slog.Int("request_id", int(header.RequestID)),
			slog.Int("response_id", int(resHeader.RequestID)),
			slog.Int("response_to", int(resHeader.ResponseTo)),
		)

		return nil, nil, lazyerrors.Errorf(
			"response_to is not equal to request_id (response_to=%d; expected=%d)",
			resHeader.ResponseTo,
			header.RequestID,
		)
	}

	if m, ok := resBody.(*wire.OpMsg); ok {
		if m.Flags != 0 {
			return nil, nil, lazyerrors.Errorf("unsupported response flags %s", m.Flags)
		}
	}

	return resHeader, resBody, nil
}
