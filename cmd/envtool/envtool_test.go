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
	"os"
	"testing"

	zapadapter "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestPrintDiagnosticData(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		var buf bytes.Buffer
		l := testutil.Logger(t)
		err := printDiagnosticData(&buf, nil, l.Sugar())
		require.NoError(t, err)
	})
}

func TestShellMkDirRmDir(t *testing.T) {
	t.Parallel()

	t.Run("Absent", func(t *testing.T) {
		err := shellRmDir("absent")
		assert.NoError(t, err)
	})

	paths := []string{"ab/c", "ab"}

	err := shellMkDir(paths...)
	assert.NoError(t, err)

	for _, path := range paths {
		assert.DirExists(t, path)
	}

	err = shellRmDir(paths...)
	assert.NoError(t, err)

	for _, path := range paths {
		assert.NoDirExists(t, path)
	}
}

func TestShellRead(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp("", "test_read")
	assert.NoError(t, err)

	s := "test string in a file"
	_, err = f.Write([]byte(s))
	assert.NoError(t, err)

	var output bytes.Buffer
	err = shellRead(&output, f.Name())
	assert.NoError(t, err)
	assert.Equal(t, s, output.String())
}

func TestPackageVersion(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp("", "test_print_version")
	assert.NoError(t, err)

	s := "v1.0.0"
	_, err = f.Write([]byte(s))
	assert.NoError(t, err)

	var output bytes.Buffer
	err = packageVersion(&output, f.Name())
	assert.NoError(t, err)
	assert.Equal(t, "1.0.0", output.String())
}

func TestSetupPostgresUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	baseURI := "postgres://username@127.0.0.1:5432/ferretdb?search_path="
	cfg, err := pgxpool.ParseConfig(baseURI)
	require.NoError(t, err)

	l := testutil.Logger(t)
	cfg.MinConns = 0
	cfg.MaxConns = 1
	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   zapadapter.NewLogger(l),
		LogLevel: tracelog.LogLevelTrace,
	}

	ctx := testutil.Ctx(t)
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)

	dbName := testutil.DatabaseName(t)
	sanitizedName := pgx.Identifier{dbName}.Sanitize()

	_, err = p.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", sanitizedName))
	require.NoError(t, err)

	// use template0 because template1 may already have the user created
	_, err = p.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s TEMPLATE template0", sanitizedName))
	require.NoError(t, err)

	t.Cleanup(func() {
		defer p.Close()

		_, err = p.Exec(context.Background(), fmt.Sprintf("DROP DATABASE %s", sanitizedName))
		require.NoError(t, err)
	})

	uri := fmt.Sprintf("postgres://username@127.0.0.1:5432/%s", dbName)

	err = setupUser(ctx, l.Sugar(), uri)
	require.NoError(t, err)

	// if the user already exists, it should not fail
	err = setupUser(ctx, l.Sugar(), uri)
	require.NoError(t, err)
}
