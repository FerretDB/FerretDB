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

package setup

import (
	"flag"
	"strings"

	"go.uber.org/zap/zapcore"
)

// flags contains all command-line flags.
//
// It is a global struct to make it easier to track what functions use flags.
var flags struct {
	targetURL     string
	targetBackend string

	targetProxyAddr  string
	targetTLS        bool
	targetUnixSocket bool

	postgreSQLURL string
	sqliteURL     string
	hanaURL       string

	compatURL string

	benchDocs int

	logLevel   zapcore.Level
	debugSetup bool

	disableFilterPushdown bool
	enableSortPushdown    bool
	enableOplog           bool

	useNewPg   bool
	useNewHana bool

	shareServer bool
}

// allBackends is a list of all supported backends.
var allBackends = []string{"ferretdb-pg", "ferretdb-sqlite", "ferretdb-hana", "mongodb"}

// init initializes flags.
func init() {
	flag.StringVar(&flags.targetURL, "target-url", "", "target system's URL; if empty, in-process FerretDB is used")
	flag.StringVar(&flags.targetBackend, "target-backend", "", "target system's backend: '%s'"+strings.Join(allBackends, "', '"))

	flag.StringVar(&flags.targetProxyAddr, "target-proxy-addr", "", "in-process FerretDB: use given proxy")
	flag.BoolVar(&flags.targetTLS, "target-tls", false, "in-process FerretDB: use TLS")
	flag.BoolVar(&flags.targetUnixSocket, "target-unix-socket", false, "in-process FerretDB: use Unix socket")

	flag.StringVar(&flags.postgreSQLURL, "postgresql-url", "", "in-process FerretDB: PostgreSQL URL for 'pg' handler.")
	flag.StringVar(&flags.sqliteURL, "sqlite-url", "", "in-process FerretDB: SQLite URI for 'sqlite' handler.")
	flag.StringVar(&flags.hanaURL, "hana-url", "", "in-process FerretDB: Hana URL for 'hana' handler.")

	flag.StringVar(&flags.compatURL, "compat-url", "", "compat system's (MongoDB) URL for compatibility tests; if empty, they are skipped")

	flag.IntVar(&flags.benchDocs, "bench-docs", 1000, "benchmarks: number of documents to generate per iteration")

	flags.logLevel = zapcore.DebugLevel
	flag.Var(&flags.logLevel, "log-level", "log level for tests")
	flag.BoolVar(&flags.debugSetup, "debug-setup", false, "enable debug logs for tests setup")

	flag.BoolVar(&flags.disableFilterPushdown, "disable-filter-pushdown", false, "disable filter pushdown")
	flag.BoolVar(&flags.enableSortPushdown, "enable-sort-pushdown", false, "enable sort pushdown")
	flag.BoolVar(&flags.enableOplog, "enable-oplog", false, "enable OpLog")

	flag.BoolVar(&flags.useNewPg, "use-new-pg", false, "use new PostgreSQL backend")
	flag.BoolVar(&flags.useNewHana, "use-new-hana", false, "use new SAP HANA backend")

	flag.BoolVar(&flags.shareServer, "share-server", false, "make all tests share listener/handler/backend")
}
