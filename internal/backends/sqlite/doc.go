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

// Package sqlite provides SQLite backend.
//
// # Design principles
//
//  1. Transactions should be avoided when possible.
//     That's because there can be, at most, one write [transaction] at a given time for the whole database.
//     (Note that there is a separate branch of SQLite with [concurrent transactions], but it is not merged yet.)
//     FerretDB often could use more granular locks - for example, per collection.
//  2. Explicit transaction retries and [SQLITE_BUSY] handling should be avoided - see above.
//     Additionally, SQLite retries automatically with the [busy_timeout] parameter we set by default, which should be enough.
//  3. Metadata is heavily cached to avoid most queries and transactions.
//
// [transaction]: https://www.sqlite.org/lang_transaction.html
// [concurrent transactions]: https://www.sqlite.org/cgi/src/doc/begin-concurrent/doc/begin_concurrent.md
// [SQLITE_BUSY]: https://www.sqlite.org/rescode.html#busy
// [busy_timeout]: https://www.sqlite.org/pragma.html#pragma_busy_timeout
package sqlite
