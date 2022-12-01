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
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
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

var (
	//go:embed error.tmpl
	errorTemplateB []byte

	// Parsed error template.
	errorTemplate = template.Must(template.New("error").Option("missingkey=error").Parse(string(errorTemplateB)))
)

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
	pgPool, err := pgdb.NewPool(ctx, &pgdb.NewPoolOpts{
		ConnString:    connString,
		Logger:        logger.Desugar(),
		Lazy:          false,
		StateProvider: p,
		OpLog:         false,
	})
	if err != nil {
		return err
	}
	defer pgPool.Close()

	logger.Info("Creating databases...")

	for _, db := range []string{"admin", "test"} {
		err = pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
			return pgdb.CreateDatabaseIfNotExists(ctx, tx, db)
		})
		if err != nil && !errors.Is(err, pgdb.ErrAlreadyExist) {
			return err
		}
	}

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

	return nil
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

// setup runs all setup commands.
func setup(ctx context.Context, logger *zap.SugaredLogger) error {
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

// runCommand runs command with given arguments.
func runCommand(command string, args []string, stdout io.Writer, logger *zap.SugaredLogger) error {
	bin, err := exec.LookPath(command)
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, args...)
	logger.Debugf("Running %s", strings.Join(cmd.Args, " "))

	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(args, " "), err)
	}

	return nil
}

// printDiagnosticData prints diagnostic data and error template on stdout.
func printDiagnosticData(setupError error, logger *zap.SugaredLogger) {
	runCommand("docker", []string{"compose", "ps", "--all"}, os.Stdout, logger)

	runCommand("docker", []string{"compose", "logs"}, os.Stdout, logger)

	var buf bytes.Buffer

	var gitVersion string
	if err := runCommand("git", []string{"version"}, &buf, logger); err != nil {
		gitVersion = err.Error()
	} else {
		gitVersion = buf.String()
	}

	buf.Reset()

	var dockerVersion string
	if err := runCommand("docker", []string{"version"}, &buf, logger); err != nil {
		dockerVersion = err.Error()
	} else {
		dockerVersion = buf.String()
	}

	buf.Reset()

	var composeVersion string
	if err := runCommand("docker", []string{"compose", "version"}, &buf, logger); err != nil {
		composeVersion = err.Error()
	} else {
		composeVersion = buf.String()
	}

	info := version.Get()

	errorTemplate.Execute(os.Stdout, map[string]any{
		"Error": setupError,

		"GOOS":   runtime.GOOS,
		"GOARCH": runtime.GOARCH,

		"Version": info.Version,
		"Commit":  info.Commit,
		"Branch":  info.Branch,
		"Dirty":   info.Dirty,
		"Debug":   info.Debug,

		"GoVersion":      runtime.Version(),
		"GitVersion":     strings.TrimSpace(gitVersion),
		"DockerVersion":  strings.TrimSpace(dockerVersion),
		"ComposeVersion": strings.TrimSpace(composeVersion),

		"NewIssueURL": "https://github.com/FerretDB/FerretDB/issues/new/choose",
	})
}

// cli struct represents all command-line commands, fields and flags.
// It's used for parsing the user input.
var cli struct {
	Debug bool `help:"Enable debug mode."`
}

func main() {
	kong.Parse(&cli)

	// always enable debug logging on CI
	if t, _ := strconv.ParseBool(os.Getenv("CI")); t {
		cli.Debug = true
	}

	level := zap.InfoLevel
	if cli.Debug {
		level = zap.DebugLevel
	}

	logging.Setup(level, "")
	logger := zap.S()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := setup(ctx, logger); err != nil {
		printDiagnosticData(err, logger)
		os.Exit(1)
	}
}
