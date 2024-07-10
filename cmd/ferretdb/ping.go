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
	"net"
	"net/url"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// ReadyZ represents the Readiness probe, which is used to run `ping`
// command against the FerretDB instance specified by cli flags.
type ReadyZ struct {
	l *zap.Logger
}

// Probe executes ping queries to open listeners, and returns true if they succeed.
// Any errors that occure are passed through ReadyZ.l listener.
//
// It is only executed if --setup-database flag is set.
func (ready *ReadyZ) Probe(ctx context.Context) bool {
	l := ready.l

	if cli.Setup.Database == "" {
		l.Info("Setup database not specified - skipping ping.")
		return true
	}

	var urls []string

	if cli.Listen.Addr != "" {
		host, port, err := net.SplitHostPort(cli.Listen.Addr)
		if err != nil {
			l.Error("Getting host and port failed.", zap.Error(err))
			return false
		}

		l.Sugar().Debugf("--listen-addr flag is set. Ping to %s will be performed.", cli.Listen.Addr)

		if host == "" {
			host = "127.0.0.1"

			l.Sugar().Debugf("Host not specified, defaulting to %s.", host)
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
		l.Warn("TLS ping is not implemented yet.")
	}

	if cli.Listen.Unix != "" {
		l.Sugar().Debugf("--listen-unix flag is set. Ping to %s will be performed.", cli.Listen.Unix)

		urls = append(urls, "mongodb://"+url.PathEscape(cli.Listen.Unix))
	}

	if len(urls) == 0 {
		l.Info("Neither --listen-addr nor --listen-unix nor --listen-tls flags were specified - skipping ping.")
		return true
	}

	for _, u := range urls {
		l.Sugar().Debugf("Pinging %s...", u)

		ctx, cancel := context.WithTimeout(ctx, cli.Setup.Timeout)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(u))
		if err != nil {
			l.Error("Connection failed.", zap.Error(err))
			return false
		}

		pingErr := client.Ping(ctx, nil)

		// do not leave connection open when ping error causes os.Exit with Fatal
		if err = client.Disconnect(ctx); err != nil {
			l.Error("Disconnect failed.", zap.Error(err))
			return false
		}

		if pingErr != nil {
			l.Error("Ping failed.", zap.Error(pingErr))
			return false
		}

		var uri *url.URL
		if uri, err = url.Parse(u); err == nil {
			u = uri.Redacted()
		}

		l.Sugar().Infof("Ping to %s successful.", u)
	}

	return true
}
