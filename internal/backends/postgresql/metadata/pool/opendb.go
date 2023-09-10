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

package pool

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

var (
	// The only supported encoding in canonical form.
	encoding = "UTF8"

	// Supported locales in canonical forms.
	locales = []string{"POSIX", "C", "C.UTF8", "en_US.UTF8"}
)

// openDB creates a pool of connections to PostgreSQL database
// and check that it works (authentication passes, settings are okay).
func openDB(uri string, l *zap.Logger, sp *state.Provider) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// version could change without FerretDB restart
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		var v string
		var err error //nolint:vet // to avoid capturing the outer variable

		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err = conn.QueryRow(ctx, `SHOW server_version`).Scan(&v); err != nil {
			return lazyerrors.Error(err)
		}

		if sp.Get().HandlerVersion != v {
			if err = sp.Update(func(s *state.State) { s.HandlerVersion = v }); err != nil {
				l.Error("openDB: failed to update state", zap.Error(err))
			}
		}

		return nil
	}

	// TODO port logging, tracing

	// see https://github.com/jackc/pgx/issues/1726#issuecomment-1711612138
	ctx := context.TODO()

	p, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if checkConnection(ctx, p) != nil {
		p.Close()
		return nil, lazyerrors.Error(err)
	}

	return p, nil
}

// simplify simplifies PostgreSQL setting value for comparison.
func simplify(v string) string {
	return strings.ToLower(strings.ReplaceAll(v, "-", ""))
}

// checkConnection checks that connection works and PostgreSQL settings are what we expect.
func checkConnection(ctx context.Context, p *pgxpool.Pool) error {
	rows, err := p.Query(ctx, "SHOW ALL")
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		// handle variable number of columns as a workaround for https://github.com/cockroachdb/cockroach/issues/101715
		values, err := rows.Values()
		if err != nil {
			return lazyerrors.Error(err)
		}

		if len(values) < 2 {
			return lazyerrors.Errorf("invalid row: %#v", values)
		}
		n, v := values[0].(string), values[1].(string)

		switch n {
		case "server_encoding", "client_encoding":
			if simplify(v) != simplify(encoding) {
				return lazyerrors.Errorf("%q is %q; supported value is %q", n, v, encoding)
			}

		case "lc_collate", "lc_ctype":
			if !slices.ContainsFunc(locales, func(l string) bool { return simplify(v) == simplify(l) }) {
				return lazyerrors.Errorf("%q is %q; supported values are %v", n, v, locales)
			}

		case "standard_conforming_strings":
			// To sanitize safely: https://github.com/jackc/pgx/issues/868#issuecomment-725544647
			if v != "on" {
				return lazyerrors.Errorf("%q is %q, want %q", n, v, "on")
			}
		}
	}

	if err := rows.Err(); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
