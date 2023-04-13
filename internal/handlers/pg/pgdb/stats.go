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

	"github.com/jackc/pgx/v4"
)

// DBStats describes statistics for a FerretDB database (PostgreSQL schema).
type DBStats struct {
	Name             string
	CountCollections int32
	CountObjects     int32
	SizeTotal        int64
	SizeIndexes      int64
	SizeCollections  int64
	CountIndexes     int32
}

// CalculateDBStats returns statistics for the given FerretDB database.
//
// If the database does not exist, it returns an object filled with the given db name and zeros for all the other fields.
func CalculateDBStats(ctx context.Context, tx pgx.Tx, db string) (*DBStats, error) {
	res := DBStats{
		Name: db,
	}

	// Total size is the disk space used by all the relations in the given schema, including tables, indexes and TOAST data.
	// It also includes the size of FerretDB metadata relations.
	sql := `
		SELECT 
		    SUM(pg_total_relation_size(quote_ident(schemaname) || '.' || quote_ident(tablename))) 
		FROM pg_tables 
		WHERE schemaname = $1`
	args := []any{db}

	row := tx.QueryRow(ctx, sql, args...)

	var schemaSize *int64
	if err := row.Scan(&schemaSize); err != nil {
		return nil, err
	}

	if schemaSize == nil {
		// If the schema size is nil, it means the schema does not exist or empty, no need to check other stats.
		return &res, nil
	}

	res.SizeTotal = *schemaSize

	// For the rest of the stats, we need to filter out FerretDB metadata relations by its reserved prefix.
	// https://wiki.postgresql.org/wiki/Count_estimate
	// https://www.postgresql.org/docs/15/catalog-pg-class.html - reltuples (could be negative)
	sql = `
	SELECT 
	    COUNT(distinct t.tablename)                                                                     AS CountTables,
		COUNT(distinct i.indexname)                                                                     AS CountIndexes,
		COALESCE(SUM(c.reltuples), 0)                                                                   AS CountRows,
		COALESCE(SUM(pg_table_size(quote_ident(t.schemaname) || '.' || quote_ident(t.tablename))), 0) 	AS SizeTables,
		COALESCE(SUM(pg_indexes_size(quote_ident(t.schemaname) || '.' || quote_ident(t.tablename))), 0) AS SizeIndexes
	FROM pg_tables AS t
		LEFT OUTER JOIN pg_class   AS c ON c.relname = t.tablename AND c.oid = (quote_ident(t.schemaname) || '.' || quote_ident(t.tablename))::regclass
		LEFT OUTER JOIN pg_indexes AS i ON i.schemaname = t.schemaname AND i.tablename = t.tablename
	WHERE t.schemaname = $1 AND t.tablename NOT LIKE $2`
	args = []any{db, reservedPrefix + "%"}

	row = tx.QueryRow(ctx, sql, args...)
	if err := row.Scan(&res.CountCollections, &res.CountIndexes, &res.CountObjects, &res.SizeCollections, &res.SizeIndexes); err != nil {
		return nil, err
	}

	return &res, nil
}

/*
// Stats returns a set of statistics for FerretDB server, database, collection
// - or, in terms of PostgreSQL, database, schema, table.
//
// It returns ErrTableNotExist is the given collection does not exist, and ignores other errors.
func (pgPool *Pool) Stats(ctx context.Context, db, collection string) (*DBStats, error) {
	res := &DBStats{
		Name: db,
	}

	err := pgPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		// See https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-DBOBJECT

		sql := `
	SELECT COUNT(distinct t.table_name)                                                         AS CountTables,
		COALESCE(SUM(s.n_live_tup), 0)                                                          AS CountRows,
		COALESCE(SUM(pg_total_relation_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0) AS SizeTotal,
		COALESCE(SUM(pg_indexes_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0)        AS SizeIndexes,
		COALESCE(SUM(pg_relation_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0)       AS SizeRelation,
		COUNT(distinct i.indexname)                                                             AS CountIndexes
	FROM information_schema.tables AS t
		LEFT OUTER JOIN pg_stat_user_tables AS s ON s.schemaname = t.table_schema AND s.relname = t.table_name
		LEFT OUTER JOIN pg_indexes          AS i ON i.schemaname = t.table_schema AND i.tablename = t.table_name`

		// TODO Exclude service schemas from the query above https://github.com/FerretDB/FerretDB/issues/1068

		var args []any

		if db != "" {
			sql += " WHERE t.table_schema = $1"
			args = append(args, db)

			if collection != "" {
				metadata, err := newMetadataStorage(tx, db, collection).get(ctx, false)
				if err != nil {
					return err
				}

				sql += " AND t.table_name = $2"
				args = append(args, metadata.table)
			}
		}

		row := tx.QueryRow(ctx, sql, args...)

		return row.Scan(&res.CountTables, &res.CountRows, &res.SizeTotal, &res.SizeIndexes, &res.SizeRelation, &res.CountIndexes)
	})

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, ErrTableNotExist):
		// return this error as is because it can be handled by the caller
		return nil, err
	default:
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}
*/
