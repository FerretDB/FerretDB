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
// # Design principles.
//
//  1. Interfaces are relatively high-level and "fat" (or not).
//     We are generally doing one backend interface call per handler call.
//     For example, `insert` command handler calls only
//     `db.Database("database").Collection("collection").Insert(ctx, params)` method that would
//     create a database if needed, create a collection if needed, and insert all documents with correct parameters.
//     There is no method to insert one document into an existing collection.
//     That shifts some complexity from a single handler into multiple backend implementations;
//     for example, support for `insert` with `ordered: true` and `ordered: false` should be implemented multiple times.
//     But that allows those implementations to be much more effective.
//  2. Backend objects are stateful.
//     Database objects are almost stateless but should be Close()'d to avoid connection leaks.
//     Collection objects are fully stateless.
//  3. Contexts are per-operation and should not be stored.
//  4. Errors returned by methods could be nil, *Error, or some other opaque error type.
//     *Error values can't be wrapped or be present anywhere in the error chain.
//     Contracts enforce *Error codes; they are not documented in the code comments
//     but are visible in the contract's code (to avoid duplication).
//
// Figure it out, especially point number 1. Update, expand, etc.
// TODO https://github.com/FerretDB/FerretDB/issues/3069
package backends
