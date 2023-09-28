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
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

var (
	//go:embed error.tmpl
	errorTemplateB []byte

	// Parsed error template.
	errorTemplate = template.Must(template.New("error").Option("missingkey=error").Parse(string(errorTemplateB)))
)

// versionFile contains version information with leading v.
const versionFile = "build/version/version.txt"

// waitForPort waits for the given port to be available until ctx is canceled.
func waitForPort(ctx context.Context, logger *zap.SugaredLogger, port uint16) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	logger.Infof("Waiting for %s to be up...", addr)

	var retry int64
	for ctx.Err() == nil {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return nil
		}

		logger.Infof("%s: %s", addr, err)

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	return fmt.Errorf("failed to connect to %s", addr)
}

// setupAnyPostgres configures given PostgreSQL.
func setupAnyPostgres(ctx context.Context, logger *zap.SugaredLogger, uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	port, err := strconv.ParseUint(u.Port(), 10, 16)
	if err != nil {
		return err
	}

	if err = waitForPort(ctx, logger, uint16(port)); err != nil {
		return err
	}

	sp, err := state.NewProvider("")
	if err != nil {
		return err
	}

	var pgPool *pgdb.Pool

	var retry int64
	for ctx.Err() == nil {
		if pgPool, err = pgdb.NewPool(ctx, uri, logger.Desugar(), sp); err == nil {
			break
		}

		logger.Infof("%s: %s", uri, err)

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	defer pgPool.Close()

	logger.Info("Creating databases...")

	for _, name := range []string{"admin", "test"} {
		err = pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
			return pgdb.CreateDatabaseIfNotExists(ctx, tx, name)
		})
		if err != nil && !errors.Is(err, pgdb.ErrAlreadyExist) {
			return err
		}
	}

	logger.Info("Tweaking settings...")

	return pgPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		for _, q := range []string{
			`CREATE ROLE readonly NOINHERIT LOGIN PASSWORD 'readonly_password'`,

			// TODO Grant permissions to readonly role.
			// https://github.com/FerretDB/FerretDB/issues/1025

			`ANALYZE`, // to make tests more stable
		} {
			if _, err = tx.Exec(ctx, q); err != nil {
				return err
			}
		}

		return nil
	})
}

// setupPostgres configures `postgres` container.
func setupPostgres(ctx context.Context, logger *zap.SugaredLogger) error {
	// user `username` must exist, but password may be any, even empty
	return setupAnyPostgres(ctx, logger.Named("postgres"), "postgres://username@127.0.0.1:5432/ferretdb")
}

// setupPostgresSecured configures `postgres_secured` container.
func setupPostgresSecured(ctx context.Context, logger *zap.SugaredLogger) error {
	return setupAnyPostgres(ctx, logger.Named("postgres_secured"), "postgres://username:password@127.0.0.1:5433/ferretdb")
}

// setupMongodb configures `mongodb` container.
func setupMongodb(ctx context.Context, logger *zap.SugaredLogger) error {
	if err := waitForPort(ctx, logger.Named("mongodb"), 47017); err != nil {
		return err
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3310
	// eval := `'rs.initiate({_id: "mongodb-rs", members: [{_id: 0, host: "localhost:47017" }]})'`
	eval := `db.serverStatus()`
	args := []string{"compose", "exec", "-T", "mongodb", "mongosh", "--port=47017", "--eval", eval}

	var buf bytes.Buffer
	var retry int64

	for ctx.Err() == nil {
		buf.Reset()

		err := runCommand("docker", args, &buf, logger)
		if err == nil {
			break
		}

		logger.Infof("%s:\n%s", err, buf.String())

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	return ctx.Err()
}

// setupMongodbSecured configures `mongodb_secured` container.
func setupMongodbSecured(ctx context.Context, logger *zap.SugaredLogger) error {
	return waitForPort(ctx, logger.Named("mongodb_secured"), 47018)
}

// setup runs all setup commands.
func setup(ctx context.Context, logger *zap.SugaredLogger) error {
	go debug.RunHandler(ctx, "127.0.0.1:8089", prometheus.DefaultRegisterer, logger.Named("debug").Desugar())

	for _, f := range []func(context.Context, *zap.SugaredLogger) error{
		setupPostgres,
		setupPostgresSecured,
		setupMongodb,
		setupMongodbSecured,
	} {
		if err := f(ctx, logger); err != nil {
			return err
		}
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

	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	logger.Debugf("Running %s", strings.Join(cmd.Args, " "))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(cmd.Args, " "), err)
	}

	return nil
}

// printDiagnosticData prints diagnostic data and error template on stdout.
func printDiagnosticData(setupError error, logger *zap.SugaredLogger) {
	runCommand("docker", []string{"compose", "logs"}, os.Stdout, logger)

	runCommand("docker", []string{"compose", "ps", "--all"}, os.Stdout, logger)

	runCommand("docker", []string{"stats", "--all", "--no-stream"}, os.Stdout, logger)

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

		"Version":    info.Version,
		"Commit":     info.Commit,
		"Branch":     info.Branch,
		"Dirty":      info.Dirty,
		"Package":    info.Package,
		"DebugBuild": info.DebugBuild,

		"GoVersion":      runtime.Version(),
		"GitVersion":     strings.TrimSpace(gitVersion),
		"DockerVersion":  strings.TrimSpace(dockerVersion),
		"ComposeVersion": strings.TrimSpace(composeVersion),

		"NewIssueURL": "https://github.com/FerretDB/FerretDB/issues/new/choose",
	})
}

// shellMkDir creates all directories from given paths.
func shellMkDir(paths ...string) error {
	var errs error

	for _, path := range paths {
		if err := os.MkdirAll(path, 0o777); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// shellRmDir removes all directories from given paths.
func shellRmDir(paths ...string) error {
	var errs error

	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// shellRead will show the content of a file.
func shellRead(w io.Writer, paths ...string) error {
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fmt.Fprint(w, string(b))
	}

	return nil
}

// packageVersion will print out FerretDB's package version (omitting leading v).
func packageVersion(w io.Writer, file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	v := string(b)
	v = strings.TrimPrefix(v, "v")

	_, err = fmt.Fprint(w, v)

	return err
}

// cli struct represents all command-line commands, fields and flags.
// It's used for parsing the user input.
//
//nolint:vet // for readability
var cli struct {
	Debug bool `help:"Enable debug mode."`

	Setup struct{} `cmd:"" help:"Setup development environment."`

	PackageVersion struct{} `cmd:"" help:"Print package version."`

	Shell struct {
		Mkdir struct {
			Paths []string `arg:"" name:"path" help:"Paths to create." type:"path"`
		} `cmd:"" help:"Create directories if they do not already exist."`
		Rmdir struct {
			Paths []string `arg:"" name:"path" help:"Paths to remove." type:"path"`
		} `cmd:"" help:"Remove directories."`
		Read struct {
			Paths []string `arg:"" name:"path" help:"Paths to read." type:"path"`
		} `cmd:"" help:"Read files."`
	} `cmd:""`

	Tests struct {
		Run struct {
			ShardIndex uint   `help:"Shard index, starting from 1."`
			ShardTotal uint   `help:"Total number of shards."`
			Run        string `help:"Run only tests matching the regexp."`

			Args []string `arg:"" help:"Other arguments and flags for 'go test'." passthrough:""`
		} `cmd:"" help:"Run tests."`
	} `cmd:""`

	Fuzz struct {
		Corpus struct {
			Src string `arg:"" help:"Source, one of: 'seed', 'generated', or collected corpus' directory."`
			Dst string `arg:"" help:"Destination, one of: 'seed', 'generated', or collected corpus' directory."`
		} `cmd:"" help:"Sync fuzz corpora."`
	} `cmd:""`
}

func main() {
	kongCtx := kong.Parse(&cli)

	// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
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

	cmd := kongCtx.Command()
	logger.Debugf("Command: %q", cmd)

	var err error

	switch cmd {
	case "setup":
		err = setup(ctx, logger)

	case "package-version":
		err = packageVersion(os.Stdout, versionFile)

	case "shell mkdir <path>":
		err = shellMkDir(cli.Shell.Mkdir.Paths...)
	case "shell rmdir <path>":
		err = shellRmDir(cli.Shell.Rmdir.Paths...)
	case "shell read <path>":
		err = shellRead(os.Stdout, cli.Shell.Read.Paths...)

	case "tests run <args>":
		err = testsRun(os.Stdout, cli.Tests.Run.ShardIndex, cli.Tests.Run.ShardTotal, cli.Tests.Run.Run, cli.Tests.Run.Args)

	case "fuzz corpus <src> <dst>":
		var seedCorpus, generatedCorpus string

		if seedCorpus, err = os.Getwd(); err != nil {
			logger.Fatal(err)
		}

		if generatedCorpus, err = fuzzGeneratedCorpus(); err != nil {
			logger.Fatal(err)
		}

		var src, dst string

		switch cli.Fuzz.Corpus.Src {
		case "seed":
			src = seedCorpus
		case "generated":
			src = generatedCorpus
		default:
			if src, err = filepath.Abs(cli.Fuzz.Corpus.Src); err != nil {
				logger.Fatal(err)
			}
		}

		switch cli.Fuzz.Corpus.Dst {
		case "seed":
			// Because we would need to add `/testdata/fuzz` back, and that's not very easy.
			logger.Fatal("Copying to seed corpus is not supported.")
		case "generated":
			dst = generatedCorpus
		default:
			dst, err = filepath.Abs(cli.Fuzz.Corpus.Dst)
			if err != nil {
				logger.Fatal(err)
			}
		}

		err = fuzzCopyCorpus(src, dst, logger)

	default:
		err = fmt.Errorf("unknown command: %s", cmd)
	}

	if err != nil {
		if cmd == "setup" {
			printDiagnosticData(err, logger)
		}
		os.Exit(1)
	}
}
