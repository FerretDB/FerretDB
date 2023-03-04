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

package hanadb

import (
	"context"
	"database/sql"
	"fmt"

	_ "SAP/go-hdb/driver" //nolint:goimports // goimports and gci are conflicting
	"go.uber.org/zap"
)

// Pool represents SAP HANA concurrency-safe connection pool.
type Pool struct {
	*sql.DB
}

// NewPool returns a new concurrency-safe connection pool.
func NewPool(ctx context.Context, url string, logger *zap.Logger) (*Pool, error) {
	pool, err := sql.Open("hdb", url)
	if err != nil {
		return nil, fmt.Errorf("hanadb.NewPool: %w", err)
	}

	// Check connection
	err = pool.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("hanadb.NewPool: %w", err)
	}

	res := &Pool{
		DB: pool,
	}

	return res, nil
}
