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

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/types"
)

func TestSortDocuments(t *testing.T) {
	t.Parallel()

	type args struct {
		docs []*types.Document
		sort *types.Document
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
		sorted  []*types.Document
	}{
		{
			name: "CompareStrings",
			args: args{
				docs: []*types.Document{
					types.MustNewDocument("_id", int32(1), "name", "Central Park Cafe", "borough", "Manhattan"),
					types.MustNewDocument("_id", int32(2), "name", "Rock A Feller Bar and Grill", "borough", "Queens"),
					types.MustNewDocument("_id", int32(3), "name", "Empire State Pub", "borough", "Brooklyn"),
					types.MustNewDocument("_id", int32(4), "name", "Stan's Pizzaria", "borough", "Manhattan"),
					types.MustNewDocument("_id", int32(5), "name", "Jane's Deli", "borough", "Brooklyn"),
				},
				sort: types.MustNewDocument("borough", int32(1)),
			},
			sorted: []*types.Document{
				types.MustNewDocument("_id", int32(3), "name", "Empire State Pub", "borough", "Brooklyn"),
				types.MustNewDocument("_id", int32(5), "name", "Jane's Deli", "borough", "Brooklyn"),
				types.MustNewDocument("_id", int32(1), "name", "Central Park Cafe", "borough", "Manhattan"),
				types.MustNewDocument("_id", int32(4), "name", "Stan's Pizzaria", "borough", "Manhattan"),
				types.MustNewDocument("_id", int32(2), "name", "Rock A Feller Bar and Grill", "borough", "Queens"),
			},
		},
		{
			name: "CompareInt",
			args: args{
				docs: []*types.Document{
					types.MustNewDocument("_id", int32(1), "building", int32(10)),
					types.MustNewDocument("_id", int32(2), "building", int32(2)),
					types.MustNewDocument("_id", int32(3), "building", int32(15)),
					types.MustNewDocument("_id", int32(4), "building", int32(1)),
					types.MustNewDocument("_id", int32(5), "building", int32(5)),
				},
				sort: types.MustNewDocument("building", int32(1)),
			},
			sorted: []*types.Document{
				types.MustNewDocument("_id", int32(4), "building", int32(1)),
				types.MustNewDocument("_id", int32(2), "building", int32(2)),
				types.MustNewDocument("_id", int32(5), "building", int32(5)),
				types.MustNewDocument("_id", int32(1), "building", int32(10)),
				types.MustNewDocument("_id", int32(3), "building", int32(15)),
			},
		},
		{
			name: "CompareStringsAndName",
			args: args{
				docs: []*types.Document{
					types.MustNewDocument("_id", int32(1), "name", "Central Park Cafe", "borough", "Manhattan"),
					types.MustNewDocument("_id", int32(2), "name", "Rock A Feller Bar and Grill", "borough", "Queens"),
					types.MustNewDocument("_id", int32(3), "name", "Empire State Pub", "borough", "Brooklyn"),
					types.MustNewDocument("_id", int32(4), "name", "Stan's Pizzaria", "borough", "Manhattan"),
					types.MustNewDocument("_id", int32(5), "name", "Jane's Deli", "borough", "Brooklyn"),
				},
				sort: types.MustNewDocument("borough", int32(1), "name", int32(1)),
			},
			sorted: []*types.Document{
				types.MustNewDocument("_id", int32(3), "name", "Empire State Pub", "borough", "Brooklyn"),
				types.MustNewDocument("_id", int32(5), "name", "Jane's Deli", "borough", "Brooklyn"),
				types.MustNewDocument("_id", int32(1), "name", "Central Park Cafe", "borough", "Manhattan"),
				types.MustNewDocument("_id", int32(4), "name", "Stan's Pizzaria", "borough", "Manhattan"),
				types.MustNewDocument("_id", int32(2), "name", "Rock A Feller Bar and Grill", "borough", "Queens"),
			},
		},
		{
			name: "CompareStringsAndID",
			args: args{
				docs: []*types.Document{
					types.MustNewDocument("_id", int32(1), "name", "Central Park Cafe", "borough", "Manhattan"),
					types.MustNewDocument("_id", int32(2), "name", "Rock A Feller Bar and Grill", "borough", "Queens"),
					types.MustNewDocument("_id", int32(3), "name", "Empire State Pub", "borough", "Brooklyn"),
					types.MustNewDocument("_id", int32(4), "name", "Stan's Pizzaria", "borough", "Manhattan"),
					types.MustNewDocument("_id", int32(5), "name", "Jane's Deli", "borough", "Brooklyn"),
				},
				sort: types.MustNewDocument("borough", int32(1), "_id", int32(1)),
			},
			sorted: []*types.Document{
				types.MustNewDocument("_id", int32(3), "name", "Empire State Pub", "borough", "Brooklyn"),
				types.MustNewDocument("_id", int32(5), "name", "Jane's Deli", "borough", "Brooklyn"),
				types.MustNewDocument("_id", int32(1), "name", "Central Park Cafe", "borough", "Manhattan"),
				types.MustNewDocument("_id", int32(4), "name", "Stan's Pizzaria", "borough", "Manhattan"),
				types.MustNewDocument("_id", int32(2), "name", "Rock A Feller Bar and Grill", "borough", "Queens"),
			},
		},
		{
			name: "SortOrderReverse",
			args: args{
				docs: []*types.Document{
					types.MustNewDocument("_id", int32(1), "name", "Central Park Cafe", "borough", "Manhattan"),
					types.MustNewDocument("_id", int32(2), "name", "Rock A Feller Bar and Grill", "borough", "Queens"),
					types.MustNewDocument("_id", int32(3), "name", "Empire State Pub", "borough", "Brooklyn"),
					types.MustNewDocument("_id", int32(4), "name", "Stan's Pizzaria", "borough", "Manhattan"),
					types.MustNewDocument("_id", int32(5), "name", "Jane's Deli", "borough", "Brooklyn"),
				},
				sort: types.MustNewDocument("_id", int32(-1)),
			},
			sorted: []*types.Document{
				types.MustNewDocument("_id", int32(5), "name", "Jane's Deli", "borough", "Brooklyn"),
				types.MustNewDocument("_id", int32(4), "name", "Stan's Pizzaria", "borough", "Manhattan"),
				types.MustNewDocument("_id", int32(3), "name", "Empire State Pub", "borough", "Brooklyn"),
				types.MustNewDocument("_id", int32(2), "name", "Rock A Feller Bar and Grill", "borough", "Queens"),
				types.MustNewDocument("_id", int32(1), "name", "Central Park Cafe", "borough", "Manhattan"),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := SortDocuments(tc.args.docs, tc.args.sort)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.sorted, tc.args.docs)
		})
	}
}
