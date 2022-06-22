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

package ferretdb

import (
	"context"
	"fmt"
	"time"
)

// ExampleRun is a testable example for Run func.
func ExampleRun() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	conf := Config{ConnectionString: "postgres://postgres@127.0.0.1:5432/ferretdb"}

	fdb := New(conf)
	connStr, _ := fdb.Run(ctx, conf)

	cancel()
	fmt.Println(connStr)
	// Output: mongodb://127.0.0.1:27017
}
