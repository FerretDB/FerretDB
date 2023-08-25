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

package pg

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/backends"
)

// collection implements backends.Collection interface.
type collection struct {
	dbName string
	name   string
}

// newCollection creates a new Collection.
func newCollection(dbName, name string) backends.Collection {
	return backends.CollectionContract(&collection{
		dbName: dbName,
		name:   name,
	})
}

// Query implements backends.Collection interface.
func (c *collection) Query(ctx context.Context, params *backends.QueryParams) (*backends.QueryResult, error) {
	panic("not implemented")
}

// Insert implements backends.Collection interface.
func (c *collection) Insert(ctx context.Context, params *backends.InsertParams) (*backends.InsertResult, error) {
	panic("not implemented")
}

// Update implements backends.Collection interface.
func (c *collection) Update(ctx context.Context, params *backends.UpdateParams) (*backends.UpdateResult, error) {
	panic("not implemented")
}

// Delete implements backends.Collection interface.
func (c *collection) Delete(ctx context.Context, params *backends.DeleteParams) (*backends.DeleteResult, error) {
	panic("not implemented")
}

// Explain implements backends.Collection interface.
func (c *collection) Explain(ctx context.Context, params *backends.ExplainParams) (*backends.ExplainResult, error) {
	panic("not implemented")
}

// check interfaces
var (
	_ backends.Collection = (*collection)(nil)
)
