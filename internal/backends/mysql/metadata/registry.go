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
type Registry struct{}
