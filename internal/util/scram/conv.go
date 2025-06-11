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

package scram

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"sync"

	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Conv represents a server SCRAM conversation.
//
// Conversation is not restartable. A new instance should be created for each conversation.
type Conv struct {
	// The order of fields is weird to make the struct smaller due to alignment.
	// All fields except l are protected by rw.

	clientFirst *message
	serverFirst *message
	clientFinal *message
	serverFinal *message
	l           *slog.Logger
	rw          sync.RWMutex
}

// NewConv creates a server SCRAM conversation.
func NewConv(l *slog.Logger) *Conv {
	return &Conv{
		l: l,
	}
}

// Succeed returns true if conversation was done successfully.
func (c *Conv) Succeed() bool {
	if c == nil {
		return false
	}

	c.rw.RLock()
	defer c.rw.RUnlock()

	return c.serverFinal != nil
}

// Username returns client's identification.
// It might not be authenticated.
func (c *Conv) Username() string {
	if c == nil {
		return ""
	}

	c.rw.RLock()
	defer c.rw.RUnlock()

	if c.clientFirst != nil {
		return c.clientFirst.n
	}

	return ""
}

// ClientFirst processes the client-first message and returns the username.
func (c *Conv) ClientFirst(payload string) (string, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	if c.clientFirst != nil {
		return "", lazyerrors.New("client-first message already processed")
	}

	var err error

	if c.clientFirst, err = parseMessage(payload, c.l); err != nil {
		return "", lazyerrors.Error(err)
	}

	if !c.clientFirst.isClientFirst() {
		return "", lazyerrors.New("unexpected client-first message")
	}

	return c.clientFirst.n, nil
}

// ServerFirst processes the ScramSha256GetSaltAndIterations's result and returns the server-first message.
func (c *Conv) ServerFirst(res wirebson.RawDocument) (string, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	if c.serverFirst != nil {
		return "", lazyerrors.New("server-first message already processed")
	}

	resDoc, err := res.Decode()
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	if resDoc.Get("ok") != int32(1) {
		return "", lazyerrors.New("unexpected response: " + resDoc.LogMessageIndent())
	}

	iterations, _ := resDoc.Get("iterations").(int32)
	if iterations == 0 {
		return "", lazyerrors.New("unexpected response: " + resDoc.LogMessageIndent())
	}

	salt, _ := resDoc.Get("salt").(string)
	if salt == "" {
		return "", lazyerrors.New("unexpected response: " + resDoc.LogMessageIndent())
	}

	// Nonce size is not specified by RFC; use the same length as the client.
	// Minimal size is already checked by [parseMessage].
	r := make([]byte, base64.StdEncoding.DecodedLen(len(c.clientFirst.r)))
	if _, err = rand.Read(r); err != nil {
		return "", lazyerrors.Error(err)
	}

	c.serverFirst = &message{
		r: c.clientFirst.r + base64.StdEncoding.EncodeToString(r),
		s: salt,
		i: int(iterations),
	}
	must.BeTrue(c.serverFirst.isServerFirst())

	return c.serverFirst.String(), nil
}

// ClientFinal processes the client-final message and returns the auth message and the client proof.
func (c *Conv) ClientFinal(payload string) (string, string, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	if c.clientFinal != nil {
		return "", "", lazyerrors.New("client-final message already processed")
	}

	var err error

	if c.clientFinal, err = parseMessage(payload, c.l); err != nil {
		return "", "", lazyerrors.Error(err)
	}

	if !c.clientFinal.isClientFinal() {
		return "", "", lazyerrors.New("unexpected client-final message")
	}

	c.clientFirst.gs2 = ""

	p := c.clientFinal.p
	c.clientFinal.p = ""

	authMessage := c.clientFirst.String() + "," + c.serverFirst.String() + "," + c.clientFinal.String()

	return authMessage, p, nil
}

// ServerFinal processes the AuthenticateWithScramSha256's result and returns the server-final message.
func (c *Conv) ServerFinal(res wirebson.RawDocument) (string, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	if c.serverFinal != nil {
		return "", lazyerrors.New("server-final message already processed")
	}

	resDoc, err := res.Decode()
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	if resDoc.Get("ok") != int32(1) {
		return "", lazyerrors.New("unexpected response: " + resDoc.LogMessageIndent())
	}

	serverSignature, _ := resDoc.Get("ServerSignature").(string)
	if serverSignature == "" {
		return "", lazyerrors.New("unexpected response: " + resDoc.LogMessageIndent())
	}

	c.serverFinal = &message{
		v: serverSignature,
	}
	must.BeTrue(c.serverFinal.isServerFinal())

	return c.serverFinal.String(), nil
}
