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

package server

import (
	"encoding/json"
	"net/http"

	"github.com/FerretDB/FerretDB/v2/internal/dataapi/api"
)

var (
	// No authentication method was provided.
	errorNoAuthenticationSpecified = api.Error{
		Error:     "no authentication methods were specified",
		ErrorCode: "InvalidParameter",
	}

	// There's a missing authentication parameter.
	errorMissingAuthenticationParameter = api.Error{
		Error: "must specify some form of authentication " +
			"(either email+password, api-key, or jwt) in the request header or body",
		ErrorCode: "MissingParameter",
	}
)

// writeError encodes [api.Error] into JSON and writes it to w
// with provided HTTP status code.
func writeError(w http.ResponseWriter, err api.Error, code int) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")

	// ignore error, as writing to connection may fail if the client disconnects
	_ = json.NewEncoder(w).Encode(err)
}
