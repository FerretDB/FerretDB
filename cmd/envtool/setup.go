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

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"time"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/debug"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

// setupMongoDB configures MongoDB containers.
func setupMongoDB(ctx context.Context, logger *slog.Logger, uri, name string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return lazyerrors.Error(err)
	}

	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return lazyerrors.Error(err)
	}

	logger = logging.WithName(logger, name)

	eval := fmt.Sprintf(`'rs.initiate({_id: "rs0", members: [{_id: 0, host: "localhost:%s" }]})'`, port)
	args := []string{"compose", "exec", "-T", name, "mongosh", "--eval", eval, uri}

	var buf bytes.Buffer
	var attempt int64

	for ctx.Err() == nil {
		buf.Reset()

		err := runCommand("docker", args, &buf, logger)
		if err == nil {
			break
		}

		logger.InfoContext(ctx, fmt.Sprintf("%s:\n%s", err, buf.String()))

		attempt++
		ctxutil.SleepWithJitter(ctx, time.Second, attempt)
	}

	return ctx.Err()
}

// setupYugabyte configures yugabyte containers by creating username:password credential, because
// the user created upon docker container startup cannot authenticate with mongodb uri.
// It waits for the port to be available before creating the user.
func setupYugabyte(ctx context.Context, uri string, l *slog.Logger) error {
	if err := waitForPort(ctx, 5433, l); err != nil {
		return lazyerrors.Error(err)
	}

	sp, err := state.NewProvider("")
	if err != nil {
		return lazyerrors.Error(err)
	}

	// many error level logs are expected until the extension is available
	doNotLog := slog.New(slog.NewJSONHandler(io.Discard, nil))

	pool, err := documentdb.NewPool(uri, l, sp)
	if err != nil {
		return lazyerrors.Error(err)
	}

	defer pool.Close()

	var retry int64

	for ctx.Err() == nil {
		err = pool.WithConn(func(conn *pgx.Conn) error {
			_, err = documentdb_api.BinaryExtendedVersion(ctx, conn, doNotLog)
			return err
		})

		if err == nil {
			break
		}

		l.InfoContext(ctx, "Waiting documentdb extension to be installed", logging.Error(err))

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	if err != nil {
		return lazyerrors.Error(err)
	}

	spec := must.NotFail(wirebson.MustDocument(
		"createUser", "username",
		"pwd", "password",
		"roles", wirebson.MustArray(
			wirebson.MustDocument(
				"role", "clusterAdmin",
				"db", "admin",
			),
			wirebson.MustDocument(
				"role", "readWriteAnyDatabase",
				"db", "admin",
			),
		),
	).Encode())

	err = pool.WithConn(func(conn *pgx.Conn) error {
		_, e := documentdb_api.CreateUser(ctx, conn, l, spec)
		return e
	})
	if err != nil {
		return lazyerrors.Error(err)
	}

	l.InfoContext(ctx, "User created")

	return nil
}

// waitForPort waits for the given port to be available until ctx is canceled.
func waitForPort(ctx context.Context, port uint16, l *slog.Logger) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	l.InfoContext(ctx, "Waiting for addr to be up", slog.String("addr", addr))

	var retry int64

	for ctx.Err() == nil {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			if err = conn.Close(); err != nil {
				return lazyerrors.Error(err)
			}

			return nil
		}

		l.InfoContext(ctx, "Connecting", slog.String("addr", addr), logging.Error(err))

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	return fmt.Errorf("failed to connect to %s", addr)
}

// setup runs all setup commands.
func setup(ctx context.Context, l *slog.Logger) error {
	h, err := debug.Listen(&debug.ListenOpts{
		TCPAddr: "127.0.0.1:8089",
		L:       logging.WithName(l, "debug"),
		R:       prometheus.DefaultRegisterer,
	})
	if err != nil {
		return lazyerrors.Error(err)
	}

	go h.Serve(ctx)

	if err = setupMongoDB(ctx, l, "mongodb://username:password@127.0.0.1:47017/", "mongodb-secure"); err != nil {
		return lazyerrors.Error(err)
	}

	yugabyteURI := "postgres://yb-user:yb-pass@127.0.0.1:5433/yugabyte"
	if err = setupYugabyte(ctx, yugabyteURI, logging.WithName(l, "yugabyte")); err != nil {
		return lazyerrors.Error(err)
	}

	l.InfoContext(ctx, "Done")
	return nil
}
