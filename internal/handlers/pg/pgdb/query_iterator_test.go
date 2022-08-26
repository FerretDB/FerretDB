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

package pgdb

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

type mockLog struct{}

func (m mockLog) Log(_ context.Context, _ pgx.LogLevel, _ string, _ map[string]interface{}) {}

type mockQueryRow struct {
	idx         int
	buffer      []*types.Document
	scanErr     bool
	err         bool
	ctxCancel   context.CancelFunc
	cancelOnIdx int
}

func (m *mockQueryRow) Close() {}

func (m *mockQueryRow) Next() bool {
	return m.idx <= len(m.buffer)
}

func (m *mockQueryRow) Err() error {
	if m.err {
		return errors.New("rows err")
	}

	return nil
}

func (m *mockQueryRow) Scan(v ...interface{}) error {
	e := errors.New("scan fail")
	if m.scanErr {
		return e
	}

	if !m.Next() {
		return e
	}

	doc := m.buffer[m.idx]

	m.idx++
	if m.cancelOnIdx == m.idx && m.ctxCancel != nil {
		m.ctxCancel()
	}

	b2, err := fjson.Marshal(doc)
	if err != nil {
		panic(err)
	}

	b := v[0].(*[]byte)
	*b = b2

	return nil
}

func TestQueryIteratorDocumentsFiltered(t *testing.T) {
	type fields struct {
		logger  pgx.Logger
		rows    func(context.CancelFunc) queryRowsIterator
		sp      SQLParam
		hasNext bool
		started bool
	}
	type args struct {
		filter *types.Document
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		assert  func(*testing.T, []*types.Document)
		wantErr bool
	}{
		{
			name: "should be return a empty result when rows is nil",
			fields: fields{
				rows: func(_ context.CancelFunc) queryRowsIterator {
					return nil
				},
			},
			args: args{
				filter: genDoc("id", "1"),
			},
			assert: func(t *testing.T, got []*types.Document) {
				require.Empty(t, got)
			},
			wantErr: false,
		},
		{
			name: "should be return only documents that matched filter",
			fields: fields{
				hasNext: true,
				rows: func(_ context.CancelFunc) queryRowsIterator {
					return &mockQueryRow{
						buffer: []*types.Document{
							genDoc("id", "1"),
							genDoc("id", "2"),
							genDoc("id", "3"),
						},
					}
				},
			},
			args: args{
				filter: genDoc("id", "1"),
			},
			assert: func(t *testing.T, got []*types.Document) {
				require.Len(t, got, 1)
			},
			wantErr: false,
		},
		{
			name: "should be return empty result when not matched filter",
			fields: fields{
				hasNext: true,
				rows: func(_ context.CancelFunc) queryRowsIterator {
					return &mockQueryRow{
						buffer: []*types.Document{
							genDoc("id", "1"),
							genDoc("id", "2"),
							genDoc("id", "3"),
						},
					}
				},
			},
			args: args{
				filter: genDoc("id", "5"),
			},
			assert: func(t *testing.T, got []*types.Document) {
				require.Empty(t, got)
			},
			wantErr: false,
		},
		{
			name: "should be return error when rows.Err returns err",
			fields: fields{
				hasNext: true,
				rows: func(_ context.CancelFunc) queryRowsIterator {
					return &mockQueryRow{
						err: true,
						buffer: []*types.Document{
							genDoc("id", "1"),
							genDoc("id", "2"),
							genDoc("id", "3"),
						},
					}
				},
			},
			args: args{
				filter: genDoc("id", "1"),
			},
			assert: func(t *testing.T, got []*types.Document) {
				require.Nil(t, got)
			},
			wantErr: true,
		},
		{
			name: "should be return error when rows.Scan returns err",
			fields: fields{
				hasNext: true,
				rows: func(_ context.CancelFunc) queryRowsIterator {
					return &mockQueryRow{
						scanErr: true,
						buffer: []*types.Document{
							genDoc("id", "1"),
						},
					}
				},
			},
			args: args{
				filter: genDoc("id", "1"),
			},
			assert: func(t *testing.T, got []*types.Document) {
				require.Nil(t, got)
			},
			wantErr: true,
		},
		{
			name: "should be stop scan when context is cancelled and return documents",
			fields: fields{
				hasNext: true,
				logger:  mockLog{},
				rows: func(cancel context.CancelFunc) queryRowsIterator {
					return &mockQueryRow{
						ctxCancel:   cancel,
						cancelOnIdx: 1,
						buffer: []*types.Document{
							genDoc("id", "1"),
							genDoc("id", "2"),
							genDoc("id", "3"),
						},
					}
				},
			},
			args: args{
				filter: genDoc("id", "1"),
			},
			assert: func(t *testing.T, got []*types.Document) {
				require.Len(t, got, 1)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			qi := &queryIterator{
				logger:  tt.fields.logger,
				rows:    tt.fields.rows(cancel),
				ctx:     ctx,
				sp:      tt.fields.sp,
				hasNext: tt.fields.hasNext,
				started: tt.fields.started,
			}

			got, err := qi.DocumentsFiltered(tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("queryIterator.DocumentsFiltered() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tt.assert(t, got)
		})
	}
}

func genDoc(k, v string) *types.Document {
	return must.NotFail(types.NewDocument(k, v))
}
