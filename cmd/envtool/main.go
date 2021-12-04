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
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
)

var (
	composeBin string

	collections = []string{
		"actor",
		"address",
		"category",
		"city",
		"country",
		"customer",
		"film_actor",
		"film_category",
		"film",
		"inventory",
		"language",
		"rental",
		"staff",
		"store",
	}
)

func runCompose(args []string, stdin io.Reader, logger *zap.SugaredLogger) {
	cmd := exec.Command(composeBin, args...)
	logger.Debugf("Running %s", strings.Join(cmd.Args, " "))

	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Fatalf("%s failed: %s", strings.Join(cmd.Args, " "), err)
	}
}

func waitForPort(ctx context.Context, port uint16) error {
	for ctx.Err() == nil {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			conn.Close()

			// FIXME https://github.com/FerretDB/FerretDB/issues/92
			time.Sleep(time.Second)

			return nil
		}

		sleepCtx, sleepCancel := context.WithTimeout(ctx, time.Second)
		<-sleepCtx.Done()
		sleepCancel()
	}

	return ctx.Err()
}

func setupMongoDB(ctx context.Context) {
	start := time.Now()
	logger := zap.S().Named("mongodb")

	logger.Infof("Waiting for port 37017 to be up...")
	if err := waitForPort(ctx, 37017); err != nil {
		logger.Fatal(err)
	}

	logger.Infof("Importing database...")

	var wg sync.WaitGroup

	for _, c := range collections {
		args := fmt.Sprintf(
			`exec -T mongodb mongoimport --uri mongodb://127.0.0.1:27017/monila `+
				`--drop --maintainInsertionOrder --collection %[1]s /test_db/%[1]s.json`,
			c,
		)

		wg.Add(1)
		go func() {
			defer wg.Done()
			runCompose(strings.Split(args, " "), nil, logger)
		}()
	}

	wg.Wait()

	logger.Infof("Done in %s.", time.Since(start))
}

func setupPagila(ctx context.Context) {
	start := time.Now()
	logger := zap.S().Named("postgres.pagila")

	logger.Infof("Waiting for port 5432 to be up...")
	if err := waitForPort(ctx, 5432); err != nil {
		logger.Fatal(err)
	}

	logger.Infof("Importing database...")

	args := strings.Split(`exec -T postgres psql -U postgres -d ferretdb --quiet -f /test_db/01-pagila-schema.sql`, " ")
	runCompose(args, nil, logger)

	args = strings.Split(`exec -T postgres psql -U postgres -d ferretdb --quiet -f /test_db/02-pagila-data.sql`, " ")
	runCompose(args, nil, logger)

	args = strings.Split(`exec -T postgres psql -U postgres -d ferretdb --quiet`, " ")
	stdin := strings.NewReader(`ALTER SCHEMA public RENAME TO pagila;`)
	runCompose(args, stdin, logger)

	logger.Infof("Done in %s.", time.Since(start))
}

func setupMonila(ctx context.Context) {
	start := time.Now()
	logger := zap.S().Named("postgres.monila")

	logger.Infof("Waiting for port 5432 to be up...")
	if err := waitForPort(ctx, 5432); err != nil {
		logger.Fatal(err)
	}

	logger.Infof("Importing database...")

	args := strings.Split(`exec -T postgres psql -U postgres -d ferretdb`, " ")
	stdin := strings.NewReader(strings.Join([]string{
		`CREATE SCHEMA monila;`,
		`CREATE SCHEMA test;`,
	}, "\n"))
	runCompose(args, stdin, logger)

	pgPool, err := pg.NewPool("postgres://postgres@127.0.0.1:5432/ferretdb", logger.Desugar(), false)
	if err != nil {
		logger.Fatal(err)
	}

	// listen on all interfaces to make mongoimport below work from inside Docker
	addr := ":27018"
	if runtime.GOOS == "darwin" {
		// do not trigger macOS firewall; it works with Docker Desktop
		addr = "127.0.0.1:27018"
	}

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr: addr,
		Mode:       "normal",
		PgPool:     pgPool,
		Logger:     logger.Named("listener").Desugar(),
	})

	lCtx, lCancel := context.WithCancel(ctx)
	lDone := make(chan struct{})
	go func() {
		defer close(lDone)
		l.Run(lCtx)
	}()

	var wg sync.WaitGroup

	for _, c := range collections {
		args := strings.Split(fmt.Sprintf(
			`exec -T mongodb mongoimport --uri mongodb://host.docker.internal:27018/monila `+
				`--drop --maintainInsertionOrder --collection %[1]s /test_db/%[1]s.json`,
			c), " ")

		wg.Add(1)
		go func() {
			defer wg.Done()
			runCompose(args, nil, logger)
		}()
	}

	wg.Wait()

	lCancel()
	<-lDone

	logger.Infof("Done in %s.", time.Since(start))
}

func main() {
	logging.Setup(zap.InfoLevel)
	logger := zap.S()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go debug.RunHandler(ctx, "127.0.0.1:8089", logger.Named("debug").Desugar())

	var err error
	if composeBin, err = exec.LookPath("docker-compose"); err != nil {
		logger.Fatal(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		setupMongoDB(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		setupPagila(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		setupMonila(ctx)
	}()

	wg.Wait()

	logger.Info("Done.")
}
