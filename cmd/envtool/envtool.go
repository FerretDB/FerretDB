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
	"log/slog"
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
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/build/version"
	mysqlpool "github.com/FerretDB/FerretDB/internal/backends/mysql/metadata/pool"
	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata/pool"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/debug"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
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
func waitForPort(ctx context.Context, logger *slog.Logger, port uint16) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	logger.InfoContext(ctx, fmt.Sprintf("Waiting for %s to be up", addr))

	var retry int64
	for ctx.Err() == nil {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return nil
		}

		logger.InfoContext(ctx, fmt.Sprintf("%s: %s", addr, err))

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	return fmt.Errorf("failed to connect to %s", addr)
}

// setupAnyPostgres configures given PostgreSQL.
func setupAnyPostgres(ctx context.Context, logger *slog.Logger, uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	port, err := strconv.ParseUint(u.Port(), 10, 16)
	if err != nil {
		return err
	}

	if u.User == nil {
		return lazyerrors.New("No username specified")
	}

	if err = waitForPort(ctx, logger, uint16(port)); err != nil {
		return err
	}

	sp, err := state.NewProvider("")
	if err != nil {
		return err
	}

	p, err := pool.New(uri, logger, sp)
	if err != nil {
		return err
	}

	defer p.Close()

	username := u.User.Username()
	password, _ := u.User.Password()

	var retry int64
	for ctx.Err() == nil {
		if _, err = p.Get(username, password); err == nil {
			break
		}

		logger.InfoContext(ctx, fmt.Sprintf("%s: %s", uri, err))

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

// setupPostgres configures `postgres` container.
func setupPostgres(ctx context.Context, logger *slog.Logger) error {
	// user `username` must exist, but password may be any, even empty
	return setupAnyPostgres(ctx, logging.WithName(logger, "postgres"), "postgres://username@127.0.0.1:5432/ferretdb")
}

// setupPostgresSecured configures `postgres_secured` container.
func setupPostgresSecured(ctx context.Context, logger *slog.Logger) error {
	return setupAnyPostgres(ctx, logging.WithName(logger, "postgres_secured"), "postgres://username:password@127.0.0.1:5433/ferretdb") //nolint:lll // for readability
}

// setupMySQL configures `mysql` container.
func setupMySQL(ctx context.Context, logger *slog.Logger) error {
	uri := "mysql://root:password@127.0.0.1:3306/ferretdb"

	sp, err := state.NewProvider("")
	if err != nil {
		return err
	}

	if err = waitForPort(ctx, logging.WithName(logger, "mysql"), 3306); err != nil {
		return err
	}

	p, err := mysqlpool.New(uri, logger, sp)
	if err != nil {
		return err
	}

	defer p.Close()

	var retry int64
	for ctx.Err() == nil {
		db, err := p.Get("root", "password")
		if err == nil {
			if _, err = db.ExecContext(ctx, "GRANT ALL PRIVILEGES ON *.* TO 'username'@'%';"); err != nil {
				return lazyerrors.Error(err)
			}

			break
		}

		logger.InfoContext(ctx, fmt.Sprintf("%s: %s", uri, err))

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

// setupMongodb configures `mongodb` container.
func setupMongodb(ctx context.Context, logger *slog.Logger) error {
	if err := waitForPort(ctx, logging.WithName(logger, "mongodb"), 47017); err != nil {
		return err
	}

	eval := `'rs.initiate({_id: "rs0", members: [{_id: 0, host: "localhost:47017" }]})'`
	args := []string{"compose", "exec", "-T", "mongodb", "mongosh", "--port=47017", "--eval", eval}

	var buf bytes.Buffer
	var retry int64

	for ctx.Err() == nil {
		buf.Reset()

		err := runCommand("docker", args, &buf, logger)
		if err == nil {
			break
		}

		logger.InfoContext(ctx, fmt.Sprintf("%s:\n%s", err, buf.String()))

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	return ctx.Err()
}

// setupMongodbSecured configures `mongodb_secured` container.
func setupMongodbSecured(ctx context.Context, logger *slog.Logger) error {
	if err := waitForPort(ctx, logging.WithName(logger, "mongodb_secured"), 47018); err != nil {
		return err
	}

	eval := `'rs.initiate({_id: "rs0", members: [{_id: 0, host: "localhost:47018" }]})'`
	shell := `mongodb://username:password@127.0.0.1:47018/?tls=true&tlsCertificateKeyFile=/etc/certs/client.pem&tlsCaFile=/etc/certs/rootCA-cert.pem` //nolint:lll // for readability
	args := []string{"compose", "exec", "-T", "mongodb_secured", "mongosh", "--eval", eval, "--shell", shell}

	var buf bytes.Buffer
	var retry int64

	for ctx.Err() == nil {
		buf.Reset()

		err := runCommand("docker", args, &buf, logger)
		if err == nil {
			break
		}

		logger.InfoContext(ctx, fmt.Sprintf("%s:\n%s", err, buf.String()))

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	return ctx.Err()
}

// setup runs all setup commands.
func setup(ctx context.Context, logger *slog.Logger) error {
	h, err := debug.Listen(&debug.ListenOpts{
		TCPAddr: "127.0.0.1:8089",
		L:       logging.WithName(logger, "debug"),
		R:       prometheus.DefaultRegisterer,
	})
	if err != nil {
		return lazyerrors.Error(err)
	}

	go h.Serve(ctx)

	for _, f := range []func(context.Context, *slog.Logger) error{
		setupPostgres,
		setupPostgresSecured,
		setupMySQL,
		setupMongodb,
		setupMongodbSecured,
	} {
		if err := f(ctx, logger); err != nil {
			return err
		}
	}

	logger.InfoContext(ctx, "Done")
	return nil
}

// runCommand runs command with given arguments.
func runCommand(command string, args []string, stdout io.Writer, logger *slog.Logger) error {
	bin, err := exec.LookPath(command)
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, args...)
	logger.Debug(fmt.Sprintf("Running %s", strings.Join(cmd.Args, " ")))

	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %s", strings.Join(cmd.Args, " "), err)
	}

	return nil
}

// printDiagnosticData prints diagnostic data and error template on stdout.
func printDiagnosticData(w io.Writer, setupError error, logger *slog.Logger) error {
	_ = runCommand("docker", []string{"compose", "logs"}, w, logger)

	_ = runCommand("docker", []string{"compose", "ps", "--all"}, w, logger)

	_ = runCommand("docker", []string{"stats", "--all", "--no-stream"}, w, logger)

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

	return errorTemplate.Execute(w, map[string]any{
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

// TestsRunParams represents `envtool tests run` parameters.
//
//nolint:vet // for readability
type TestsRunParams struct {
	ShardIndex uint   `help:"Shard index, starting from 1."`
	ShardTotal uint   `help:"Total number of shards."`
	Run        string `help:"Run only tests matching the regexp."`
	Skip       string `help:"Skip tests matching the regexp."`

	Args []string `arg:"" help:"Other arguments and flags for 'go test'." passthrough:""`
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
		Run TestsRunParams `cmd:"" help:"Run tests."`
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

	level := slog.LevelInfo
	if cli.Debug {
		level = slog.LevelDebug
	}

	opts := &logging.NewHandlerOpts{
		Base:         "console",
		Level:        level,
		RemoveTime:   true,
		RemoveSource: true,
	}

	logging.Setup(opts, "")
	logger := slog.Default()

	cmd := kongCtx.Command()
	logger.Debug(fmt.Sprintf("Command: %q", cmd))

	var err error

	switch cmd {
	case "setup":
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

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
		ctx, stop := ctxutil.SigTerm(context.Background())
		defer stop()

		err = testsRun(ctx, &cli.Tests.Run, logger)

	case "fuzz corpus <src> <dst>":
		var seedCorpus, generatedCorpus string

		if seedCorpus, err = os.Getwd(); err != nil {
			logger.LogAttrs(context.Background(), logging.LevelFatal, "Failed to get current directory", logging.Error(err))
		}

		if generatedCorpus, err = fuzzGeneratedCorpus(); err != nil {
			logger.LogAttrs(context.Background(), logging.LevelFatal, "Failed to generate fuzz corpus", logging.Error(err))
		}

		var src, dst string

		switch cli.Fuzz.Corpus.Src {
		case "seed":
			src = seedCorpus
		case "generated":
			src = generatedCorpus
		default:
			if src, err = filepath.Abs(cli.Fuzz.Corpus.Src); err != nil {
				logger.LogAttrs(context.Background(), logging.LevelFatal, "Unknown fuzz corpus source", logging.Error(err))
			}
		}

		switch cli.Fuzz.Corpus.Dst {
		case "seed":
			// Because we would need to add `/testdata/fuzz` back, and that's not very easy.
			logger.LogAttrs(
				context.Background(),
				logging.LevelFatal,
				"Copying to seed corpus is not supported",
				logging.Error(err),
			)
		case "generated":
			dst = generatedCorpus
		default:
			dst, err = filepath.Abs(cli.Fuzz.Corpus.Dst)
			if err != nil {
				logger.LogAttrs(context.Background(), logging.LevelFatal, "Unknown fuzz corpus destination", logging.Error(err))
			}
		}

		err = fuzzCopyCorpus(src, dst, logger)

	default:
		err = fmt.Errorf("unknown command: %s", cmd)
	}

	if err != nil {
		if cmd == "setup" {
			_ = printDiagnosticData(os.Stderr, err, logger)
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			logger.Error("Failed to exit", logging.Error(exitErr))
			os.Exit(exitErr.ExitCode())
		}

		logger.LogAttrs(context.Background(), logging.LevelFatal, "Failed unexpectedly", logging.Error(err))
	}
}
