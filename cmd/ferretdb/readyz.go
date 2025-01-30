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
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// ReadyZ represents the Readiness probe.
//
// It connects to FerretDB with Go driver and sends `ping` command.
type ReadyZ struct {
	l *slog.Logger
}

// Probe implements [debug.Probe].
func (r *ReadyZ) Probe(ctx context.Context) bool {
	var urls []string

	if cli.Listen.Addr != "" {
		host, port, err := net.SplitHostPort(cli.Listen.Addr)
		if err != nil {
			r.l.ErrorContext(ctx, "Getting host and port failed", logging.Error(err))
			return false
		}

		if host == "" {
			host = "127.0.0.1"
		}

		u := &url.URL{
			Scheme: "mongodb",
			Host:   net.JoinHostPort(host, port),
			Path:   "/",
		}

		urls = append(urls, u.String())
	}

	if cli.Listen.TLS != "" {
		// TODO https://github.com/FerretDB/FerretDB/issues/4427
		r.l.WarnContext(ctx, "TLS ping is not implemented yet")
	}

	if cli.Listen.Unix != "" {
		urls = append(urls, "mongodb://"+url.PathEscape(cli.Listen.Unix))
	}

	if len(urls) == 0 {
		r.l.InfoContext(ctx, "Nothing to ping")
		return true
	}

	for _, u := range urls {
		r.l.DebugContext(ctx, fmt.Sprintf("Pinging %s", u))

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(u))
		if err != nil {
			r.l.ErrorContext(ctx, "Connection failed", logging.Error(err))
			return false
		}

		pingErr := client.Ping(ctx, nil)

		err = client.Disconnect(ctx)

		if pingErr != nil {
			r.l.ErrorContext(ctx, "Ping failed", logging.Error(pingErr))
			return false
		}

		if err != nil {
			r.l.ErrorContext(ctx, "Disconnect failed", logging.Error(err))
			return false
		}

		r.l.InfoContext(ctx, fmt.Sprintf("Ping to %s successful", u))
	}

	return true
}
