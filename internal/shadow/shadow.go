// Copyright 2021 Baltoro OÜ.
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

package shadow

import (
	"bufio"
	"context"
	"net"

	"github.com/MangoDB-io/MangoDB/internal/wire"
)

type Handler struct {
	bufr *bufio.Reader
	bufw *bufio.Writer
}

func New(addr string) (*Handler, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Handler{
		bufr: bufio.NewReader(conn),
		bufw: bufio.NewWriter(conn),
	}, nil
}

func (c *Handler) Handle(ctx context.Context, header *wire.MsgHeader, msg wire.MsgBody) (*wire.MsgHeader, wire.MsgBody, error) {
	if err := wire.WriteMessage(c.bufw, header, msg); err != nil {
		return nil, nil, err
	}

	if err := c.bufw.Flush(); err != nil {
		return nil, nil, err
	}

	return wire.ReadMessage(c.bufr)
}
