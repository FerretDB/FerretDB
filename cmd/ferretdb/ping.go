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
	"net"
	"net/url"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	var pingExecuted bool

	if cli.Listen.Addr != "" {
		pingExecuted = true

		host, port, err := net.SplitHostPort(cli.Listen.Addr)
		if err != nil {
			l.Fatal(err)
		}

		l.Infof("--listen-addr flag is set. Pinging %s...", cli.Listen.Addr)

		if host == "" {
			host = "127.0.0.1"

			l.Debugf("Host not specified, defaulting to %s.", host)
		}

		u := &url.URL{
			Scheme: "mongodb",
			Host:   fmt.Sprintf("%s:%s", host, port),
			Path:   cli.Setup.Database,
			User:   url.UserPassword(cli.Setup.Username, cli.Setup.Password),
		}

		l.Debugf("Pinging %s...", u)

		ctx, cancel := context.WithTimeout(context.Background(), cli.Setup.Timeout)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(u.String()))
		if err != nil {
			l.Fatal(err)
		}

		defer func() {
			if err = client.Disconnect(ctx); err != nil {
				l.Fatal(err)
			}
		}()

		if err = client.Ping(ctx, nil); err != nil {
			l.Fatal(err)
		}

		l.Info("Ping successful.")
	}

	if cli.Listen.Unix != "" {
		pingExecuted = true

		ctx, cancel := context.WithTimeout(context.Background(), cli.Setup.Timeout)
		defer cancel()

		l.Infof("--listen-unix flag is set. Pinging %s...", cli.Listen.Unix)

		u := "mongodb://" + url.PathEscape(cli.Listen.Unix)

		l.Debugf("Pinging %s...", u)

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(u))
		if err != nil {
			l.Fatal(err)
		}

		defer func() {
			if err = client.Disconnect(ctx); err != nil {
				l.Fatal(err)
			}
		}()

		if err = client.Ping(ctx, nil); err != nil {
			l.Fatal(err)
		}

		l.Info("Ping successful.")
	}

	if !pingExecuted {
		l.Info("Neither --listen-addr nor --listen-unix flags were specified - skipping ping.")
	}
}
