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

// Package clientconn provides wire protocol server implementation.
package clientconn

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"time"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// conn represents client connection.
type conn struct {
	netConn        net.Conn
	l              *slog.Logger
	m              *middleware.Middleware
	testRecordsDir string // if empty, no records are created
}

// run runs the client connection until ctx is canceled, client disconnects,
// or fatal error or panic is encountered.
//
// Returned error is always non-nil.
//
// The caller is responsible for closing the underlying net.Conn.
func (c *conn) run(ctx context.Context) (err error) {
	ci := conninfo.New()
	ctx, cancel := context.WithCancelCause(conninfo.Ctx(ctx, ci))

	defer func() {
		if p := recover(); p != nil {
			c.l.LogAttrs(ctx, logging.LevelDPanic, fmt.Sprint(p), logging.Error(err))
			err = lazyerrors.New("panic")
		}

		cancel(lazyerrors.Errorf("run exits: %w", err))
		ci.Close()
	}()

	if c.netConn.RemoteAddr().Network() != "unix" {
		ci.Peer, err = netip.ParseAddrPort(c.netConn.RemoteAddr().String())
		if err != nil {
			err = lazyerrors.Error(err)
			return
		}
	}

	go func() {
		<-ctx.Done()

		// unblocks ReadMessage in the processRequest below; any non-zero past value will do
		_ = c.netConn.SetDeadline(time.Unix(0, 0))
	}()

	bufr := bufio.NewReader(c.netConn)

	// if test record path is set, split netConn reader to write to file and bufr
	if c.testRecordsDir != "" {
		if err = os.MkdirAll(c.testRecordsDir, 0o777); err != nil {
			return
		}

		// write to temporary file first, then rename to avoid partial files

		// use local directory so os.Rename below always works
		var f *os.File
		if f, err = os.CreateTemp(c.testRecordsDir, "_*.partial"); err != nil {
			return
		}

		h := sha256.New()

		defer func() {
			c.renamePartialFile(ctx, f, h, err)
		}()

		r := io.TeeReader(c.netConn, io.MultiWriter(f, h))
		bufr = bufio.NewReader(r)
	}

	bufw := bufio.NewWriter(c.netConn)
	requests, readerDone := readRequestStream(ctx, bufr, cancel)
	disconnectDone := watchPeerDisconnect(ctx, c.netConn, cancel)
	defer func() {
		cancel(lazyerrors.Errorf("run exits: %w", err))
		<-readerDone
		<-disconnectDone
	}()

	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case req, ok := <-requests:
			if !ok {
				return context.Cause(ctx)
			}
			if err = c.processRequest(ctx, req, bufw); err != nil {
				if cause := context.Cause(ctx); cause != nil {
					return cause
				}
				return
			}
		}
	}
}

// readRequestStream gives a single reader ownership of the connection. It reads
// ahead while a handler executes so EOF cancels blocked database work. An
// unbuffered handoff bounds read-ahead to one additional protocol message.
func readRequestStream(ctx context.Context, reader io.Reader, cancel context.CancelCauseFunc) (<-chan *middleware.Request, <-chan struct{}) {
	requests := make(chan *middleware.Request)
	done := make(chan struct{})
	buffered := bufio.NewReader(reader)
	go func() {
		defer close(done)
		defer close(requests)
		defer func() {
			if recover() != nil {
				cancel(errors.New("wire request reader panic"))
			}
		}()
		for {
			header, body, err := wire.ReadMessage(buffered)
			if err != nil {
				cancel(lazyerrors.Error(err))
				return
			}
			req, err := middleware.RequestWire(header, body)
			if err != nil {
				cancel(lazyerrors.Error(err))
				return
			}
			select {
			case requests <- req:
			case <-ctx.Done():
				return
			}
		}
	}()
	return requests, done
}

// processRequest passes a request to the middleware and writes the response.
//
// Any error returned indicates the connection should be closed.
func (c *conn) processRequest(ctx context.Context, req *middleware.Request, bufw *bufio.Writer) error {
	resp := c.m.Handle(ctx, req)
	if resp == nil {
		return lazyerrors.New("middleware returned nil response")
	}
	// OP_MSG moreToCome is used for unacknowledged writes. Execute the
	// operation normally, including rollback/error handling, but do not put
	// a reply on the socket that the client is not expecting to read.
	if msg, ok := req.WireBody().(*wire.OpMsg); ok && msg.Flags.FlagSet(wire.OpMsgMoreToCome) {
		return nil
	}

	if err := wire.WriteMessage(bufw, resp.WireHeader(), resp.WireBody()); err != nil {
		return lazyerrors.Error(err)
	}

	if err := bufw.Flush(); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// renamePartialFile takes over an open file `f` and closes it.
// It uses the given error to check if the connection was closed by the client,
// if so the given file is renamed to a name generated by hash,
// otherwise, it deletes the given file.
func (c *conn) renamePartialFile(ctx context.Context, f *os.File, h hash.Hash, err error) {
	// do not store partial files
	if !errors.Is(err, wire.ErrZeroRead) {
		_ = f.Close()
		_ = os.Remove(f.Name())

		return
	}

	// surprisingly, Sync is required before Rename on many OS/FS combinations
	if e := f.Sync(); e != nil {
		c.l.WarnContext(ctx, "Failed to sync file", logging.Error(e))
	}

	if e := f.Close(); e != nil {
		c.l.WarnContext(ctx, "Failed to close file", logging.Error(e))
	}

	fileName := hex.EncodeToString(h.Sum(nil))

	hashPath := filepath.Join(c.testRecordsDir, fileName[:2])
	if e := os.MkdirAll(hashPath, 0o777); e != nil {
		c.l.WarnContext(ctx, "Failed to make directory", logging.Error(e))
	}

	path := filepath.Join(hashPath, fileName+".bin")
	if e := os.Rename(f.Name(), path); e != nil {
		c.l.WarnContext(ctx, "Failed to rename file", logging.Error(e))
	}
}
