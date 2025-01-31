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

package logging

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgxLogger(t *testing.T) {
	t.Parallel()

	level := tracelog.LogLevelTrace

	var buf bytes.Buffer
	h := NewHandler(&buf, &NewHandlerOpts{
		Base:          "console",
		Level:         pgxLogLevels[level],
		CheckMessages: true,
	})

	tracer := &tracelog.TraceLog{
		Logger:   NewPgxLogger(slog.New(h)),
		LogLevel: level,
	}

	config, err := pgxpool.ParseConfig("postgres://invalid/")
	require.NoError(t, err)

	config.ConnConfig.Tracer = tracer

	ctx := context.TODO()

	p, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err)

	t.Cleanup(p.Close)

	err = p.Ping(ctx)
	require.Error(t, err)

	assert.Regexp(t, `failed to connect`, buf.String())
}
