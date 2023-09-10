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

// Package backends provides common interfaces and code for all backend implementations.
//
// # Design principles
//
//  1. Interfaces are designed to balance the efficiency of individual backends and code duplication between them.
//     For example, the `Collection.InsertAll` method creates a database and collection automatically if needed.
//     Theoretically, the handler can make three separate backend calls
//     (create a database if needed, create collection if needed, insert documents),
//     but that implementation would likely be less efficient due to extra roundtrips, transactions, and/or locks.
//     On the other hand, the logic of `ordered` `insert`s is only present in the handler.
//     If some backend supports the same semantics as MongoDB, we will likely add a separate option method,
//     and the handler would use that before falling back to the previous behavior.
//  2. [Backend] objects are stateful.
//     [Database] and [Collection] objects are stateless.
//  3. Backends maintain the list of databases and collections.
//     It is recommended that it does so by not querying the information_schema or equivalent often.
//  4. Contexts are per-operation and should not be stored.
//     They are used for passing authentication information via [conninfo].
//  5. Errors returned by methods could be nil, [*Error], or some other opaque error type.
//     *Error values can't be wrapped or be present anywhere in the error chain.
//     Contracts enforce error codes; they are not documented in the code comments
//     but are visible in the contract's code (to avoid duplication).
//     Methods should return different error codes only if the difference is important for the handler.
//
// Update, expand, etc.
// TODO https://github.com/FerretDB/FerretDB/issues/3069
package backends
