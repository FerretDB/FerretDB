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
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tigrisdata/tigris-client-go/config"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
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

// generatedCorpus returns $GOCACHE/fuzz/github.com/FerretDB/FerretDB,
// ensuring that this directory exists.
func generatedCorpus() (string, error) {
	b, err := exec.Command("go", "env", "GOCACHE").Output()
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	path := filepath.Join(string(bytes.TrimSpace(b)), "fuzz", "github.com", "FerretDB", "FerretDB")

	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, 0o777)
		}

		if err != nil {
			return "", lazyerrors.Error(err)
		}
	}

	return path, err
}

// collectFiles returns a map of all fuzz files in the given directory.
func collectFiles(root string, logger *zap.SugaredLogger) (map[string]struct{}, error) {
	existingFiles := make(map[string]struct{}, 1000)
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return lazyerrors.Error(err)
		}

		if info.IsDir() {
			// skip .git, etc
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// skip other files
		if _, err = hex.DecodeString(info.Name()); err != nil {
			return nil
		}

		path, err = filepath.Rel(root, path)
		if err != nil {
			return lazyerrors.Error(err)
		}
		logger.Debug(path)
		existingFiles[path] = struct{}{}
		return nil
	})

	return existingFiles, err
}

// cutTestdata returns s with "/testdata/fuzz" removed.
//
// That converts seed corpus entry like `internal/bson/testdata/fuzz/FuzzArray/HEX`
// to format used by generated and collected corpora `internal/bson/FuzzArray/HEX`.
func cutTestdata(s string) string {
	old := string(filepath.Separator) + filepath.Join("testdata", "fuzz")
	return strings.Replace(s, old, "", 1)
}

// diff returns the set of files in src that are not in dst, with and without applying `cutTestdata`.
func diff(src, dst map[string]struct{}) []string {
	res := make([]string, 0, 50)

	for p := range src {
		if _, ok := dst[p]; ok {
			continue
		}

		if _, ok := dst[cutTestdata(p)]; ok {
			continue
		}

		res = append(res, p)
	}

	sort.Strings(res)

	return res
}

// copyFile copies a file from src to dst, overwriting dst if it exists.
func copyFile(src, dst string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer srcF.Close()

	dir := filepath.Dir(dst)

	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0o777)
	}

	if err != nil {
		return lazyerrors.Error(err)
	}

	dstF, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)
	if err != nil {
		return lazyerrors.Error(err)
	}

	_, err = io.Copy(dstF, srcF)
	if closeErr := dstF.Close(); err == nil {
		err = closeErr
	}

	if err != nil {
		os.Remove(dst)
		return lazyerrors.Error(err)
	}

	return nil
}

// copyCorpus copies all new corpus files from srcRoot to dstRoot.
func copyCorpus(srcRoot, dstRoot string) {
	logger := zap.S()

	srcFiles, err := collectFiles(srcRoot, logger)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("Found %d files in src.", len(srcFiles))

	dstFiles, err := collectFiles(dstRoot, logger)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("Found %d existing files in dst.", len(dstFiles))

	files := diff(srcFiles, dstFiles)
	logger.Infof("Copying new %d files to dst.", len(files))

	for _, p := range files {
		src := filepath.Join(srcRoot, p)
		dst := cutTestdata(filepath.Join(dstRoot, p))
		logger.Debugf("%s -> %s", src, dst)

		if err := copyFile(src, dst); err != nil {
			logger.Fatal(err)
		}
	}
}

// waitForPort waits for the given port to be available until ctx is done.
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

	p, err := state.NewProvider("")
	if err != nil {
		return err
	}

	var pgPool *pgdb.Pool

	var retry int64
	for ctx.Err() == nil {
		if pgPool, err = pgdb.NewPool(ctx, uri, logger.Desugar(), p); err == nil {
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

// setupAnyTigris configures given Tigris.
func setupAnyTigris(ctx context.Context, logger *zap.SugaredLogger, port uint16) error {
	err := waitForPort(ctx, logger, port)
	if err != nil {
		return err
	}

	cfg := &config.Driver{
		URL: fmt.Sprintf("127.0.0.1:%d", port),
	}

	p, err := state.NewProvider("")
	if err != nil {
		return err
	}

	var db *tigrisdb.TigrisDB

	var retry int64
	for ctx.Err() == nil {
		if db, err = tigrisdb.New(ctx, cfg, logger.Desugar(), p); err == nil {
			break
		}

		logger.Infof("%s: %s", cfg.URL, err)

		retry++
		ctxutil.SleepWithJitter(ctx, time.Second, retry)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	defer db.Driver.Close()

	logger.Info("Creating databases...")

	for _, name := range []string{"admin", "test"} {
		if _, err = db.Driver.CreateProject(ctx, name); err != nil {
			return err
		}
	}

	return nil
}

// setupTigris configures all Tigris containers.
func setupTigris(ctx context.Context, logger *zap.SugaredLogger) error {
	logger = logger.Named("tigris")

	// See docker-compose.yml.
	for _, port := range []uint16{8081, 8091, 8092, 8093, 8094} {
		if err := setupAnyTigris(ctx, logger.Named(strconv.Itoa(int(port))), port); err != nil {
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

	if err := setupPostgresSecured(ctx, logger); err != nil {
		return err
	}

	if err := setupTigris(ctx, logger); err != nil {
		return err
	}

	if err := waitForPort(ctx, logger.Named("mongodb"), 47017); err != nil {
		return err
	}

	if err := waitForPort(ctx, logger.Named("mongodb_secure"), 47018); err != nil {
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
var cli struct {
	Debug bool     `help:"Enable debug mode."`
	Setup struct{} `cmd:"" help:"Setup development environment."`
	Shell struct {
		Mkdir struct {
			Paths []string `arg:"" name:"path" help:"Paths to create." type:"path"`
		} `cmd:"" help:"Create directories if they do not already exist."`
		Rmdir struct {
			Paths []string `arg:"" name:"path" help:"Paths to remove." type:"path"`
		} `cmd:"" help:"Remove directories."`
		Read struct {
			Paths []string `arg:"" name:"path" help:"Paths to read." type:"path"`
		} `cmd:"" help:"read files"`
	} `cmd:""`
	PackageVersion struct{} `cmd:"" help:"Print package version"`
	Tests          struct {
		Shard struct {
			Index uint `help:"Shard index, starting from 1" required:""`
			Total uint `help:"Total number of shards"       required:""`
		} `cmd:"" help:"Print sharded integration tests"`
	} `cmd:""`
	Fuzz struct {
		Corpus struct {
			Src string `arg:"" help:"Source, one of: 'seed', 'generated', or collected corpus' directory."`
			Dst string `arg:"" help:"Destination, one of: 'seed', 'generated', or collected corpus' directory."`
		} `cmd:""`
	} `cmd:""`
}

func main() {
	kongCtx := kong.Parse(&cli)
	logger := zap.S()
	var err error
	// always enable debug logging on CI
	if t, _ := strconv.ParseBool(os.Getenv("CI")); t {
		cli.Debug = true
	}

	level := zap.InfoLevel
	if cli.Debug {
		level = zap.DebugLevel
	}

	seedCorpus, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}

	generatedCorpus, err := generatedCorpus()
	if err != nil {
		logger.Fatal(err)
	}

	logging.Setup(level, "")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var src, dst string

	switch cmd := kongCtx.Command(); cmd {
	case "setup":
		err = setup(ctx, logger)
	case "shell mkdir <path>":
		err = shellMkDir(cli.Shell.Mkdir.Paths...)
	case "shell rmdir <path>":
		err = shellRmDir(cli.Shell.Rmdir.Paths...)
	case "shell read <path>":
		err = shellRead(os.Stdout, cli.Shell.Read.Paths...)
	case "package-version":
		err = packageVersion(os.Stdout, versionFile)
	case "tests shard":
		err = testsShard(os.Stdout, cli.Tests.Shard.Index, cli.Tests.Shard.Total)
	case "fuzz-corpus <src> <dst>":
		switch cli.Fuzz.Corpus.Src {
		case "seed":
			src = seedCorpus
		case "generated":
			src = generatedCorpus
		default:
			src, err = filepath.Abs(cli.Fuzz.Corpus.Src)
			if err != nil {
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
	default:
		err = fmt.Errorf("unknown command: %s", cmd)

		logger.Infof("Copying from %s to %s.", src, dst)
		copyCorpus(src, dst)
	}

	if err != nil {
		printDiagnosticData(err, logger)
		os.Exit(1)
	}
}
