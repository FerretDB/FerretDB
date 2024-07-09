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

	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
)

// ping creates connection to FerretDB instance specified by the flags, and runs `ping` command against it.
// The check is only executed if --setup-database flag is set.
func ping() {
	logger := setupLogger(cli.Log.Format, "")
	checkFlags(logger)

	l := logger.Sugar()

	if cli.Setup.Database == "" {
		l.Info("Setup database not specified - skipping ping.")
		return
	}

	var urls []string

	if cli.Listen.Addr != "" {
		host, port, err := net.SplitHostPort(cli.Listen.Addr)
		if err != nil {
			logger.Fatal("Getting host and port failed.", zap.Error(err))
		}

		l.Debugf("--listen-addr flag is set. Ping to %s will be performed.", cli.Listen.Addr)

		if host == "" {
			host = "127.0.0.1"

			l.Debugf("Host not specified, defaulting to %s.", host)
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
		host, port, err := net.SplitHostPort(cli.Listen.TLS)
		if err != nil {
			logger.Fatal("Getting host and port failed.", zap.Error(err))
		}

		l.Debugf("--listen-tls flag is set. Ping to %s will be performed.", cli.Listen.Addr)

		if host == "" {
			host = "127.0.0.1"

			l.Debugf("Host not specified, defaulting to %s.", host)
		}

		if cli.Listen.TLSKeyFile == "" || cli.Listen.TLSCaFile == "" {
			logger.Fatal("When --listen-tls is set, both --listen-tls-cert-file and --listen-tls-ca-file need to be provided.")
		}

		values := url.Values{}

		values.Add("tls", "true")
		values.Add("tlsCaFile", cli.Listen.TLSCaFile)
		values.Add("tlsCertificateKeyFile", cli.Listen.TLSKeyFile)

		u := &url.URL{
			Scheme:   "mongodb",
			Host:     net.JoinHostPort(host, port),
			Path:     cli.Setup.Database,
			User:     url.UserPassword(cli.Setup.Username, cli.Setup.Password),
			RawQuery: values.Encode(),
		}

		urls = append(urls, u.String())
	}

	if cli.Listen.Unix != "" {
		l.Debugf("--listen-unix flag is set. Ping to %s will be performed.", cli.Listen.Unix)

		urls = append(urls, "mongodb://"+url.PathEscape(cli.Listen.Unix))
	}

	if len(urls) == 0 {
		l.Info("Neither --listen-addr nor --listen-unix nor --listen-tls flags were specified - skipping ping.")
		return
	}

	for _, u := range urls {
		l.Debugf("Pinging %s...", u)

		ctx, _ := ctxutil.SigTerm(context.Background())

		ctx, cancel := context.WithTimeout(ctx, cli.Setup.Timeout)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(u))
		if err != nil {
			logger.Fatal("Connection failed.", zap.Error(err))
		}

		pingErr := client.Ping(ctx, nil)

		// do not leave connection open when ping error causes os.Exit with Fatal
		if err = client.Disconnect(ctx); err != nil {
			logger.Fatal("Disconnect failed.", zap.Error(err))
		}

		if pingErr != nil {
			logger.Fatal("Ping failed.", zap.Error(pingErr))
		}

		var uri *url.URL
		if uri, err = url.Parse(u); err == nil {
			u = uri.Redacted()
		}

		l.Infof("Ping to %s successful.", u)
	}
}
