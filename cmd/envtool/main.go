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
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/version"
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
	if err := tryCommand(composeBin, args, stdin, nil, logger); err != nil {
		logger.Fatal(err)
	}
}

func tryCommand(command string, args []string, stdin io.Reader, stdout io.Writer, logger *zap.SugaredLogger) error {
	gitBin, err := exec.LookPath(command)
	if err != nil {
		return err
	}
	cmd := exec.Command(gitBin, args...)
	logger.Debugf("Running %s", strings.Join(cmd.Args, " "))

	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	if stdout != nil {
		cmd.Stdout = stdout
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(args, " "), err)
	}

	return nil
}

func waitForPort(ctx context.Context, port uint16) error {
	for ctx.Err() == nil {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			conn.Close()

			return nil
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf("failed to connect to 127.0.0.1:%d", port)
}

func waitForPostgresPort(ctx context.Context, port uint16) error {
	logger := zap.S().Named("postgres.wait")

	for ctx.Err() == nil {
		var pgPool *pgdb.Pool
		pgPool, err := pgdb.NewPool(fmt.Sprintf("postgres://postgres@127.0.0.1:%d/ferretdb", port), logger.Desugar(), false)
		if err == nil {
			pgPool.Close()

			return nil
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf("failed to connect to 127.0.0.1:%d", port)
}

func setupMongoDB(ctx context.Context) {
	start := time.Now()
	logger := zap.S().Named("mongodb")

	logger.Infof("Importing database...")

	var wg sync.WaitGroup

	for _, c := range collections {
		args := fmt.Sprintf(
			`exec -T mongodb mongoimport --uri mongodb://127.0.0.1:27017/monila `+
				`--drop --maintainInsertionOrder --collection %[1]s /test_db/monila/%[1]s.json`,
			c,
		)

		wg.Add(1)
		go func() {
			defer wg.Done()
			runCompose(strings.Split(args, " "), nil, logger)
		}()
	}

	{
		args := `exec -T mongodb mongoimport --uri mongodb://127.0.0.1:27017/values ` +
			`--drop --maintainInsertionOrder --collection values /test_db/values/values.json`

		wg.Add(1)
		go func() {
			defer wg.Done()
			runCompose(strings.Split(args, " "), nil, logger)
		}()
	}

	wg.Wait()

	logger.Infof("Done in %s.", time.Since(start))
}

func setupMonilaAndValues(ctx context.Context, pgPool *pgdb.Pool) {
	start := time.Now()
	logger := zap.S().Named("postgres.monila_and_values")

	logger.Infof("Importing databases...")

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
		Logger:     logger.Desugar(),
	})

	prometheus.DefaultRegisterer.MustRegister(l)

	lCtx, lCancel := context.WithCancel(ctx)
	lDone := make(chan struct{})
	go func() {
		defer close(lDone)
		l.Run(lCtx)
	}()

	var wg sync.WaitGroup

	for _, c := range collections {
		cmd := fmt.Sprintf(
			`exec -T mongodb mongoimport --uri mongodb://host.docker.internal:27018/monila `+
				`--drop --maintainInsertionOrder --collection %[1]s /test_db/monila/%[1]s.json`,
			c,
		)

		wg.Add(1)
		go func() {
			defer wg.Done()
			runCompose(strings.Split(cmd, " "), nil, logger)
		}()
	}

	{
		cmd := `exec -T mongodb mongoimport --uri mongodb://host.docker.internal:27018/values ` +
			`--drop --maintainInsertionOrder --collection values /test_db/values/values.json`

		wg.Add(1)
		go func() {
			defer wg.Done()
			runCompose(strings.Split(cmd, " "), nil, logger)
		}()
	}

	wg.Wait()

	lCancel()
	<-lDone

	logger.Infof("Done in %s.", time.Since(start))
}

//nolint:forbidigo // Printf used to make diagnostic data easier to copy.
func printDiagnosticData(runError error, logger *zap.SugaredLogger) {
	buffer := bytes.NewBuffer([]byte{})
	var composeVersion string
	composeError := tryCommand(composeBin, []string{"version"}, nil, buffer, logger)
	if composeError != nil {
		composeVersion = composeError.Error()
	} else {
		composeVersion = string(buffer.Bytes())
	}
	buffer.Reset()

	var dockerVersion string
	dockerError := tryCommand("git", []string{"version"}, nil, buffer, logger)
	if dockerError != nil {
		dockerVersion = dockerError.Error()
	} else {
		dockerVersion = string(buffer.Bytes())
	}

	buffer.Reset()

	var gitVersion string
	gitError := tryCommand("git", []string{"version"}, nil, buffer, logger)
	if gitError != nil {
		gitVersion = gitError.Error()
	} else {
		gitVersion = string(buffer.Bytes())
	}

	info := version.Get()

	fmt.Printf(`Looks like something went wrong..
Please file an issue with all that information below:
	
	OS: %s
	Arch: %s
	Version: %s
	Commit: %s
	Branch: %s

	Go version: %s
	%s
	%s
	%s

	Error: %v
`,
		runtime.GOOS,
		runtime.GOARCH,
		info.Version,
		info.Commit,
		info.Branch,

		runtime.Version(),
		strings.TrimSpace(gitVersion),
		strings.TrimSpace(composeVersion),
		strings.TrimSpace(dockerVersion),

		runError,
	)
}

func setupLogger(debug bool) *zap.SugaredLogger {
	logging.Setup(zap.InfoLevel)
	if debug {
		logging.Setup(zap.DebugLevel)
	}
	logger := zap.S()
	return logger
}

func parseFlags() *bool {
	debugF := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
		fmt.Fprintln(flag.CommandLine.Output(), "no arguments expected")
		os.Exit(2)
	}
	return debugF
}

func run(ctx context.Context, logger *zap.SugaredLogger) error {
	go debug.RunHandler(ctx, "127.0.0.1:8089", logger.Named("debug").Desugar())

	var err error
	composeBin, err = exec.LookPath("docker-compose")
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	portsCtx, portsCancel := context.WithTimeout(ctx, time.Minute)
	defer portsCancel()

	var portsCheckError error

	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Info("Waiting for port 37017 to be up...")
		portsCheckError = waitForPort(portsCtx, 37017)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Info("Waiting for port 5432 to be up...")
		portsCheckError = waitForPostgresPort(portsCtx, 5432)
	}()

	wg.Wait()

	if portsCheckError != nil {
		return portsCheckError
	}

	var pgPool *pgdb.Pool
	pgPool, err = pgdb.NewPool("postgres://postgres@127.0.0.1:5432/ferretdb", logger.Desugar(), false)
	if err != nil {
		return err
	}

	for _, db := range []string{`monila`, `values`, `test`} {
		if err := pgPool.CreateSchema(ctx, db); err != nil {
			return err
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		setupMongoDB(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		setupMonilaAndValues(ctx, pgPool)
	}()

	wg.Wait()

	for _, q := range []string{
		`CREATE ROLE readonly NOINHERIT LOGIN`,
		`GRANT SELECT ON ALL TABLES IN SCHEMA monila, values, test TO readonly`,
		`GRANT USAGE ON SCHEMA monila, values, test TO readonly`,
		`ANALYZE`, // to make tests more stable
	} {
		if _, err := pgPool.Exec(ctx, q); err != nil {
			return err
		}
	}

	logger.Info("Done.")
	return nil
}

func main() {
	debugLevel := parseFlags()

	logger := setupLogger(*debugLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := run(ctx, logger)
	if err != nil {
		printDiagnosticData(err, logger)
		os.Exit(2)
	}
}
