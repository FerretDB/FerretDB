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

package api

import _ "embed"

// Spec is the embedded OpenAPI specification.
//
//go:embed openapi.json
var Spec []byte

//go:generate ../../../bin/oapi-codegen --config=./oapi-config.yml ./openapi.json

// Response types are used to represent the responses from the API.
//
//sumtype:decl
type Response interface {
	sealed()
}

func (r *AggregateResponseBody) sealed()  {}
func (r *DeleteResponseBody) sealed()     {}
func (r *Error) sealed()                  {}
func (r *FindOneResponseBody) sealed()    {}
func (r *FindManyResponseBody) sealed()   {}
func (r *InsertOneResponseBody) sealed()  {}
func (r *InsertManyResponseBody) sealed() {}
func (r *UpdateResponseBody) sealed()     {}

var (
	_ Response = (*AggregateResponseBody)(nil)
	_ Response = (*DeleteResponseBody)(nil)
	_ Response = (*Error)(nil)
	_ Response = (*FindOneResponseBody)(nil)
	_ Response = (*FindManyResponseBody)(nil)
	_ Response = (*InsertManyResponseBody)(nil)
	_ Response = (*InsertOneResponseBody)(nil)
	_ Response = (*UpdateResponseBody)(nil)
)
