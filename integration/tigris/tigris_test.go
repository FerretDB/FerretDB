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

// Package tigris provides Tigris handler.
package tigris

import (
	"context"
	"fmt"
	"testing"

	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"
	"github.com/tigrisdata/tigris-client-go/filter"

	"github.com/FerretDB/FerretDB/internal/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

var url = "127.0.0.1:8081"

func TestRead(t *testing.T) {
	ctx := context.Background()

	drv, err := driver.NewDriver(ctx, &config.Driver{URL: url})
	if err != nil {
		panic(err)
	}

	list, err := drv.ListDatabases(ctx)
	fmt.Printf("%#v", list)
	if err != nil {
		panic(err)
	}

	err = drv.CreateDatabase(ctx, "db1")
	if err != nil {
		panic(err)
	}

	doc, err := types.NewDocument(
		"_id", types.ObjectID{0x00, 0x01, 0x02, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c},
		"string", "foo",
		"int32", int32(42),
		"int64", int64(123),
		"binary", types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
	)
	if err != nil {
		panic(err)
	}
	schema, err := tjson.DocumentSchema(doc)
	if err != nil {
		panic(err)
	}
	schema.Title = "coll1"
	b := must.NotFail(schema.Marshal())
	db := drv.UseDatabase("db1")
	err = db.CreateOrUpdateCollection(ctx, "coll1", b)
	if err != nil {
		panic(err)
	}
	b, err = tjson.Marshal(doc)
	if err != nil {
		panic(err)
	}
	_, err = db.Insert(ctx, "coll1", []driver.Document{b})
	if err != nil {
		panic(err)
	}

	id := must.NotFail(doc.Get("_id")).(types.ObjectID)
	f := must.NotFail(filter.Eq("_id", tjson.ObjectID(id)).Build())
	it, err := db.Read(ctx, "coll1", f, nil)
	if err != nil {
		panic(err)
	}

	var tdoc driver.Document
	for it.Next(&tdoc) {
		fmt.Printf("%s\n", string(tdoc))
	}
	if err := it.Err(); err != nil {
		panic(err)
	}
}
