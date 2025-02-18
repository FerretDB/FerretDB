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
	"log/slog"
	"net"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/v2/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/v2/internal/util/debug"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
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

// setup runs all setup commands.
func setup(ctx context.Context, logger *slog.Logger) error {
	h, err := debug.Listen(&debug.ListenOpts{
		TCPAddr: "127.0.0.1:8089",
		L:       logging.WithName(logger, "debug"),
		R:       prometheus.DefaultRegisterer,
	})
	if err != nil {
		return lazyerrors.Error(err)
	}

	go h.Serve(ctx)

	if err = setupMongoDB(ctx, logger, "mongodb://username:password@127.0.0.1:47017/", "mongodb-secure"); err != nil {
		return lazyerrors.Error(err)
	}

	logger.InfoContext(ctx, "Done.")
	return nil
}
