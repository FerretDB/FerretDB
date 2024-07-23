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

	"github.com/FerretDB/FerretDB/internal/util/logging"
)

// ReadyZ represents the Readiness probe, which is used to run `ping`
// command against the FerretDB instance specified by cli flags.
type ReadyZ struct {
	l *slog.Logger
}

// Probe executes ping queries to open listeners, and returns true if they succeed.
// Any errors that occure are passed through ReadyZ.l listener.
//
// It is only executed if --setup-database flag is set.
func (ready *ReadyZ) Probe(ctx context.Context) bool {
	l := ready.l

	if cli.Setup.Database == "" {
		l.InfoContext(ctx, "Setup database not specified - skipping ping")
		return true
	}

	var urls []string

	if cli.Listen.Addr != "" {
		host, port, err := net.SplitHostPort(cli.Listen.Addr)
		if err != nil {
			l.ErrorContext(ctx, "Getting host and port failed", logging.Error(err))
			return false
		}

		l.DebugContext(ctx, fmt.Sprintf("--listen-addr flag is set. Ping to %s will be performed", cli.Listen.Addr))

		if host == "" {
			host = "127.0.0.1"

			l.DebugContext(ctx, fmt.Sprintf("Host not specified, defaulting to %s", host))
		}

		u := &url.URL{
			Scheme: "mongodb",
			Host:   net.JoinHostPort(host, port),
			Path:   cli.Setup.Database,
			User:   url.UserPassword(cli.Setup.Username, cli.Setup.Password),
		}

		urls = append(urls, u.String())
	}

	if cli.Listen.TLS != "" {
		// TODO https://github.com/FerretDB/FerretDB/issues/4427
		l.WarnContext(ctx, "TLS ping is not implemented yet")
	}

	if cli.Listen.Unix != "" {
		l.DebugContext(ctx, fmt.Sprintf("--listen-unix flag is set. Ping to %s will be performed", cli.Listen.Unix))

		urls = append(urls, "mongodb://"+url.PathEscape(cli.Listen.Unix))
	}

	if len(urls) == 0 {
		l.InfoContext(ctx, "Neither --listen-addr nor --listen-unix nor --listen-tls flags were specified - skipping ping")
		return true
	}

	for _, u := range urls {
		l.DebugContext(ctx, fmt.Sprintf("Pinging %s", u))

		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, cli.Setup.Timeout)

		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(u))
		if err != nil {
			l.ErrorContext(ctx, "Connection failed", logging.Error(err))
			return false
		}

		pingErr := client.Ping(ctx, nil)

		// do not leave connection open when ping error causes os.Exit with Fatal
		if err = client.Disconnect(ctx); err != nil {
			l.ErrorContext(ctx, "Disconnect failed", logging.Error(err))
			return false
		}

		if pingErr != nil {
			l.ErrorContext(ctx, "Ping failed", logging.Error(pingErr))
			return false
		}

		var uri *url.URL
		if uri, err = url.Parse(u); err == nil {
			u = uri.Redacted()
		}

		l.InfoContext(ctx, fmt.Sprintf("Ping to %s successful", u))
	}

	return true
}
