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

package common

import "github.com/FerretDB/FerretDB/internal/types"

type CountParams struct {
	DB          string          `name:"$db"`
	Collection  string          `name:"collection"`
	Collation   any             `name:"collation,unimplemented"`
	Hint        any             `name:"hint,ignored"`
	ReadConcern any             `name:"readConcern,ignored"`
	Comment     any             `name:"comment,ignored"`
	Filter      *types.Document `name:"query"`
	Skip        int64           `name:"skip"`
	Limit       int64           `name:"limit"`
}
