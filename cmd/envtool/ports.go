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
	"time"

	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// waitForPort waits for the given port to be available until ctx is done.
func waitForPort(ctx context.Context, logger *zap.SugaredLogger, port uint16) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	logger.Infof("Waiting for %s to be up...", addr)

	for ctx.Err() == nil {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return nil
		}

		logger.Infof("%s: %s", addr, err)
		ctxutil.Sleep(ctx, time.Second)
	}

	return fmt.Errorf("failed to connect to %s", addr)
}

// waitForPostgresPort waits for the given PostgreSQL port to be available until ctx is done.
func waitForPostgresPort(ctx context.Context, logger *zap.SugaredLogger, port uint16) error {
	if err := waitForPort(ctx, logger, port); err != nil {
		return err
	}

	url := fmt.Sprintf("postgres://username:password@127.0.0.1:%d/ferretdb", port)

	for ctx.Err() == nil {
		p, err := state.NewProvider("")
		if err != nil {
			return err
		}

		pgPool, err := pgdb.NewPool(ctx, url, logger.Desugar(), false, p)
		if err == nil {
			pgPool.Close()
			return nil
		}

		logger.Infof("%s: %s", url, err)
		ctxutil.Sleep(ctx, time.Second)
	}

	return fmt.Errorf("failed to connect to %s", url)
}

// waitForTigrisPort waits for the given Tigris port to be available until ctx is done.
func waitForTigrisPort(ctx context.Context, logger *zap.SugaredLogger, port uint16) error {
	if err := waitForPort(ctx, logger, port); err != nil {
		return err
	}

	cfg := &config.Driver{
		URL: fmt.Sprintf("127.0.0.1:%d", port),
	}

	for ctx.Err() == nil {
		driver, err := driver.NewDriver(ctx, cfg)
		if err == nil {
			_, err = driver.Info(ctx)
			_ = driver.Close()

			if err == nil {
				return nil
			}
		}

		logger.Infof("%s: %s", cfg.URL, err)
		ctxutil.Sleep(ctx, time.Second)
	}

	return fmt.Errorf("failed to connect to %s", cfg.URL)
}
