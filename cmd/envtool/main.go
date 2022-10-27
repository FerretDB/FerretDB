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
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/jackc/pgx/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/version"
)

// runCommand runs command with given arguments.
func runCommand(command string, args []string, stdout io.Writer, logger *zap.SugaredLogger) error {
	bin, err := exec.LookPath(command)
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, args...)
	logger.Debugf("Running %s", strings.Join(cmd.Args, " "))

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

// waitForPort waits for the given port to be available until ctx is done.
func waitForPort(ctx context.Context, logger *zap.SugaredLogger, port uint16) error {
	logger.Infof("Waiting for port %d to be up...", port)

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

// waitForPostgresPort waits for the given PostgreSQL port to be available until ctx is done.
func waitForPostgresPort(ctx context.Context, logger *zap.SugaredLogger, port uint16) error {
	if err := waitForPort(ctx, logger, port); err != nil {
		return err
	}

	for ctx.Err() == nil {
		p, err := state.NewProvider("")
		if err != nil {
			return err
		}

		connString := fmt.Sprintf("postgres://postgres@127.0.0.1:%d/ferretdb", port)
		pgPool, err := pgdb.NewPool(ctx, connString, logger.Desugar(), false, p)
		if err == nil {
			pgPool.Close()
			return nil
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf("failed to connect to 127.0.0.1:%d", port)
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
			driver.Close()
			if err == nil {
				return nil
			}
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf("failed to connect to 127.0.0.1:%d", port)
}

//nolint:forbidigo // Printf used to make diagnostic data easier to copy.
func printDiagnosticData(runError error, logger *zap.SugaredLogger) {
	buffer := bytes.NewBuffer([]byte{})
	var composeVersion string
	composeError := runCommand("docker-compose", []string{"version"}, buffer, logger)
	if composeError != nil {
		composeVersion = composeError.Error()
	} else {
		composeVersion = buffer.String()
	}
	buffer.Reset()

	var dockerVersion string
	dockerError := runCommand("docker", []string{"version"}, buffer, logger)
	if dockerError != nil {
		dockerVersion = dockerError.Error()
	} else {
		dockerVersion = buffer.String()
	}

	buffer.Reset()

	var gitVersion string
	gitError := runCommand("git", []string{"version"}, buffer, logger)
	if gitError != nil {
		gitVersion = gitError.Error()
	} else {
		gitVersion = buffer.String()
	}

	info := version.Get()

	fmt.Printf(`Looks like something went wrong.
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

// setupPostgres configures PostgreSQL.
func setupPostgres(ctx context.Context, logger *zap.SugaredLogger) error {
	logger = logger.Named("postgres")

	if err := waitForPostgresPort(ctx, logger, 5432); err != nil {
		return err
	}

	p, err := state.NewProvider("")
	if err != nil {
		return err
	}

	connString := "postgres://postgres@127.0.0.1:5432/ferretdb"
	pgPool, err := pgdb.NewPool(ctx, connString, logger.Desugar(), false, p)
	if err != nil {
		return err
	}
	defer pgPool.Close()

	logger.Info("Tweaking settings...")

	for _, q := range []string{
		`CREATE ROLE readonly NOINHERIT LOGIN`,

		// TODO Grant permissions to readonly role.
		// https://github.com/FerretDB/FerretDB/issues/1025

		`ANALYZE`, // to make tests more stable
	} {
		if _, err = pgPool.Exec(ctx, q); err != nil {
			return err
		}
	}

	logger.Info("Creating databases...")

	err = pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		for _, db := range []string{"admin", "test"} {
			if err = pgdb.CreateDatabaseIfNotExists(ctx, tx, db); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// setupTigris configures Tigris.
func setupTigris(ctx context.Context, logger *zap.SugaredLogger) error {
	logger = logger.Named("tigris")

	if err := waitForTigrisPort(ctx, logger, 8081); err != nil {
		return err
	}

	cfg := &config.Driver{
		URL: "127.0.0.1:8081",
	}
	driver, err := driver.NewDriver(ctx, cfg)
	if err != nil {
		return err
	}
	defer driver.Close()

	logger.Info("Creating databases...")
	for _, db := range []string{"admin", "test"} {
		if err = driver.CreateDatabase(ctx, db); err != nil {
			return err
		}
	}

	return nil
}

// run runs all setup commands.
func run(ctx context.Context, logger *zap.SugaredLogger) error {
	go debug.RunHandler(ctx, "127.0.0.1:8089", prometheus.DefaultRegisterer, logger.Named("debug").Desugar())

	if err := setupPostgres(ctx, logger); err != nil {
		return err
	}

	if err := setupTigris(ctx, logger); err != nil {
		return err
	}

	if err := waitForPort(ctx, logger, 37017); err != nil {
		return err
	}

	logger.Info("Done.")
	return nil
}

// cli struct represents all command-line commands, fields and flags.
// It's used for parsing the user input.
var cli struct {
	Debug bool `help:"Enable debug mode."`
}

func main() {
	kong.Parse(&cli)

	level := zap.InfoLevel
	if cli.Debug {
		level = zap.DebugLevel
	}

	logging.Setup(level, "")
	logger := zap.S()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err := run(ctx, logger)
	if err != nil {
		printDiagnosticData(err, logger)
		os.Exit(2)
	}
}
