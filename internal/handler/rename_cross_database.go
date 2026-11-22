package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

// renameAcrossDatabases implements the plain-collection move used by UniFi
// restore. The opt-in path is pinned to the inspected DocumentDB catalog version.
// Metadata, UUID replacement, target drop and cache invalidation commit in one
// PostgreSQL transaction. Data tables, indexes and validators remain attached to
// the source collection ID; no BSON conversion or data copy occurs.
func (h *Handler) renameAcrossDatabases(ctx context.Context, oldDB, oldName, newDB, newName string, dropTarget bool) error {
	if oldDB == "local" || newDB == "local" {
		return mongoerrors.New(mongoerrors.ErrIllegalOperation, "Cannot rename between local and another database")
	}
	conn, err := h.p.AcquireContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	return pgx.BeginTxFunc(ctx, conn.Conn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		var version string
		if err := tx.QueryRow(ctx, "SELECT extversion FROM pg_extension WHERE extname='documentdb'").Scan(&version); err != nil {
			return err
		}
		if version != "0.107-0" {
			return mongoerrors.New(mongoerrors.ErrNotImplemented, "Cross-database rename requires the validated DocumentDB catalog version")
		}
		if _, err := tx.Exec(ctx, "LOCK TABLE documentdb_api_catalog.collections IN EXCLUSIVE MODE"); err != nil {
			return err
		}
		var id int64
		var view, sharded bool
		err := tx.QueryRow(ctx, `SELECT collection_id, view_definition IS NOT NULL, shard_key IS NOT NULL FROM documentdb_api_catalog.collections WHERE database_name=$1 AND collection_name=$2 FOR UPDATE`, oldDB, oldName).Scan(&id, &view, &sharded)
		if errors.Is(err, pgx.ErrNoRows) {
			return mongoerrors.New(mongoerrors.ErrNamespaceNotFound, "Source collection does not exist")
		}
		if err != nil {
			return err
		}
		if view {
			return mongoerrors.New(mongoerrors.ErrCommandNotSupportedOnView, "Cannot rename a view")
		}
		if sharded {
			return mongoerrors.New(mongoerrors.ErrNotImplemented, "Cross-database rename of sharded collections is not supported")
		}
		var targetExists bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM documentdb_api_catalog.collections WHERE database_name=$1 AND collection_name=$2)", newDB, newName).Scan(&targetExists); err != nil {
			return err
		}
		if targetExists && !dropTarget {
			return mongoerrors.New(mongoerrors.ErrNamespaceExists, "Target collection already exists")
		}
		h.p.CloseNamespaceCursors(ctx, oldDB, oldName)
		h.p.CloseNamespaceCursors(ctx, newDB, newName)
		if _, err := tx.Exec(ctx, "LOCK TABLE "+pgx.Identifier{"documentdb_data", fmt.Sprintf("documents_%d", id)}.Sanitize()+" IN ACCESS EXCLUSIVE MODE"); err != nil {
			return err
		}
		if targetExists {
			if _, err := documentdb_api.DropCollection(ctx, tx.Conn(), h.L, newDB, newName, nil, nil, false); err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `UPDATE documentdb_api_catalog.collections SET database_name=$1, collection_name=$2, collection_uuid=gen_random_uuid() WHERE collection_id=$3`, newDB, newName, id)
		return err
	})
}
