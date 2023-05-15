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
//  1. Contexts are per-operation and should not be stored.
//  2. Backend is stateful. Database and Collection are stateless.
//  3. Returned errors could be nil, *Error or any other error type.
//     *Error codes are enforced by contracts;
//     they are not documented in the code comments, but are visible in the contract's code (to avoid duplication).
package backends
