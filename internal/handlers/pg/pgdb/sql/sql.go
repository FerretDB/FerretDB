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

package sql

import (
	"context"
	"embed"
	"io/fs"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// All SQL files.
//
//go:embed *.sql
var files embed.FS

func Install(ctx context.Context, tx pgx.Tx, l *zap.Logger) error {
	return fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return lazyerrors.Error(err)
		}

		if d.IsDir() {
			return nil
		}

		b, err := files.ReadFile(path)
		if err != nil {
			return lazyerrors.Error(err)
		}

		l.Info("Executing SQL file", zap.String("path", path))

		if _, err = tx.Exec(ctx, string(b)); err != nil {
			return lazyerrors.Error(err)
		}

		return nil
	})
}
