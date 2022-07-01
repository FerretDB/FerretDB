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
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/version"
)

var composeBin string

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
		pgPool, err := pgdb.NewPool(ctx, fmt.Sprintf("postgres://postgres@127.0.0.1:%d/ferretdb", port), logger.Desugar(), false)
		if err == nil {
			pgPool.Close()

			return nil
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf("failed to connect to 127.0.0.1:%d", port)
}

//nolint:forbidigo // Printf used to make diagnostic data easier to copy.
func printDiagnosticData(runError error, logger *zap.SugaredLogger) {
	buffer := bytes.NewBuffer([]byte{})
	var composeVersion string
	composeError := tryCommand(composeBin, []string{"version"}, nil, buffer, logger)
	if composeError != nil {
		composeVersion = composeError.Error()
	} else {
		composeVersion = buffer.String()
	}
	buffer.Reset()

	var dockerVersion string
	dockerError := tryCommand("git", []string{"version"}, nil, buffer, logger)
	if dockerError != nil {
		dockerVersion = dockerError.Error()
	} else {
		dockerVersion = buffer.String()
	}

	buffer.Reset()

	var gitVersion string
	gitError := tryCommand("git", []string{"version"}, nil, buffer, logger)
	if gitError != nil {
		gitVersion = gitError.Error()
	} else {
		gitVersion = buffer.String()
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
	if composeBin, err = exec.LookPath("docker-compose"); err != nil {
		return err
	}

	logger.Info("Waiting for port MongoDB 37017 to be up...")
	if err = waitForPort(ctx, 37017); err != nil {
		return err
	}

	logger.Info("Waiting for PostgreSQL port 5432 to be up...")
	if err = waitForPostgresPort(ctx, 5432); err != nil {
		return err
	}

	pgPool, err := pgdb.NewPool(ctx, "postgres://postgres@127.0.0.1:5432/ferretdb", logger.Desugar(), false)
	if err != nil {
		return err
	}

	for _, schema := range []string{"admin", "test"} {
		if err = pgPool.CreateDatabase(ctx, schema); err != nil {
			return err
		}
	}

	for _, q := range []string{
		`CREATE ROLE readonly NOINHERIT LOGIN`,
		`GRANT SELECT ON ALL TABLES IN SCHEMA test TO readonly`,
		`GRANT USAGE ON SCHEMA test TO readonly`,
		`ANALYZE`, // to make tests more stable
	} {
		if _, err = pgPool.Exec(ctx, q); err != nil {
			return err
		}
	}

	logger.Info("Done.")
	return nil
}

func main() {
	debugLevel := parseFlags()

	logger := setupLogger(*debugLevel)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err := run(ctx, logger)
	if err != nil {
		printDiagnosticData(err, logger)
		os.Exit(2)
	}
}
