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
	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/mysql/metadata/pool"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"sync"
)

const (
	// MySQL table name where FerretDB metadata is stored.
	metadataTableName = backends.ReservedPrefix + "database_metadata"

	// MySQL max table name length.
	maxTableNameLength = 64

	// MySQL max index name length.
	maxIndexNameLength = 64
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "mysql_metadata"
)

// Registry provides access to MySQL databases and collections information.
//
// Exported methods and [getPool] are safe for concurrent use. Other unexported methods are not.
//
// All methods should call [getPool] to check authentication.
// There is no authorization yet - if username/password is correct,
// all databases and collections are visible as far as Registry is concerned.
//
// Registry metadata is loaded upon first call by client, using [conninfo] in the context of the client.
//
//nolint:vet // for readability
type Registry struct {
	p *pool.Pool
	l *zap.Logger

	// rw protects colls but also acts like a global lock for the whole registry.
	// The latter effectively replaces transactions (see the mysql backend package description for more info).
	// One global lock should be replaced by more granular locks – one per database or even one per collection.
	// But that requires some redesign.
	// TODO https://github.com/FerretDB/FerretDB/issues/2755
	rw    sync.RWMutex
	colls map[string]map[string]*Collection // database name -> collection name -> collection
}

// Close closes the registry.
func (r *Registry) Close() {
	r.p.Close()
}

// getPool returns a pool of connections to MySQL database
// for the username/password combination in the context using [conninfo]
// (or any pool if authentication is bypassed)
//
// It loads metadata if it hasn't been loaded from the database yet.
//
// It acquires read lock to check metadata, if metadata is empty it acquires write lock
// to load metadata, so it is safe for concurrent use.
//
// All methods use this method to check authentication and load metadata.
func (r *Registry) getPool(ctx context.Context) (*fsql.DB, error) {
	connInfo := conninfo.Get(ctx)

	var p *fsql.DB

	if connInfo.BypassAuth {
		if p = r.p.GetAny(); p == nil {
			return nil, lazyerrors.New("no connection pool")
		}
	} else {
		username, password := connInfo.Auth()

		var err error
		if p, err = r.p.Get(username, password); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	r.rw.RLock()
	if r.colls != nil {
		r.rw.RUnlock()
		return p, nil
	}
	r.rw.RUnlock()

	r.rw.Lock()
	defer r.rw.Unlock()

	dbNames, err := r.initDBs(ctx, p)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return p, nil
}

// initDBs returns a list of database names using schema information
func (r *Registry) initDBs(ctx context, pool *fsql.DB)

// Describe implements prometheus.Collector.
func (r *Registry) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(r, ch)
}

// Collect impements prometheus.Collector.
func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	r.p.Collect(ch)

	r.rw.RLock()
	defer r.rw.RUnlock()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "databases"),
			"The current number of database in the registry.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(len(r.colls)),
	)

	for db, colls := range r.colls {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem, "collections"),
				"The current number of collections in the registry.",
				[]string{"db"}, nil,
			),
			prometheus.GaugeValue,
			float64(len(colls)),
			db,
		)
	}
}

// check interfaces
var (
	_ prometheus.Collector = (*Registry)(nil)
)
