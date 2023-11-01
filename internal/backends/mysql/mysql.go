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

// Package sqlite provides SQLite backend.

package mysql

import "context"

// stats represents information about statistics of tables and indexes
type stats struct {
	countDocuments  int64
	sizeIndexes     int64
	sizeTables      int64
	sizeFreeStorage int64
}

// collectionsStats returns statistics about tables and indexes for the given collections.
//
// If refresh is true, it calls ANALYZE on the tables of the given list of collections.
//
// If the list of collections is empty, then stats filled with zero values is returned.
func collectionsStats(ctx context.Context) {}
