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

package metadata

import (
	"context"
	"strings"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	// The only supported encoding in canonical form.
	supportedEncoding = "UTF8"

	// Supported locales in canonical forms.
	supportedLocales = []string{"POSIX", "C", "C.UTF8", "en_US.UTF8"}
)

// simplifySetting simplifies PostgreSQL setting value for comparison.
func simplifySetting(v string) string {
	return strings.ToLower(strings.ReplaceAll(v, "-", ""))
}

// isSupportedEncoding checks `server_encoding` and `client_encoding` values.
func isSupportedEncoding(v string) bool {
	return simplifySetting(v) == simplifySetting(supportedEncoding)
}

// isSupportedLocale checks `lc_collate` and `lc_ctype` values.
func isSupportedLocale(v string) bool {
	v = simplifySetting(v)

	for _, s := range supportedLocales {
		if v == simplifySetting(s) {
			return true
		}
	}

	return false
}

// checkSettings checks PostgreSQL settings.
func checkSettings(ctx context.Context, p *pgxpool.Pool, l *zap.Logger) error {
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
		name := values[0].(string)
		value := values[1].(string)

		switch name {
		case "server_encoding", "client_encoding":
			if !isSupportedEncoding(value) {
				return lazyerrors.Errorf("%q is %q; supported value is %q", name, value, supportedEncoding)
			}
		case "lc_collate", "lc_ctype":
			if !isSupportedLocale(value) {
				return lazyerrors.Errorf("%q is %q; supported values are %v", name, value, supportedLocales)
			}
		case "standard_conforming_strings": // To sanitize safely: https://github.com/jackc/pgx/issues/868#issuecomment-725544647
			if value != "on" {
				return lazyerrors.Errorf("%q is %q, want %q", name, value, "on")
			}
		default:
			continue
		}

		l.Debug("PostgreSQL setting", zap.String(name, value))
	}

	if err := rows.Err(); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
