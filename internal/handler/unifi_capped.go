package handler

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

//go:embed unifi_capped.sql
var cappedSchemaSQL string

type cappedOptions struct{ size, maximum int64 }

func cappedNumber(value int64) any {
	if value >= math.MinInt32 && value <= math.MaxInt32 {
		return int32(value)
	}
	return value
}

func cappedScaledSize(size int64, scale float64) any {
	if scale == 1 {
		return cappedNumber(size)
	}
	if scale >= float64(math.MaxInt64) {
		return int32(0)
	}
	if scale == math.Trunc(scale) {
		return cappedNumber(size / int64(scale))
	}
	return cappedNumber(int64(math.Floor(float64(size) / scale)))
}

func cappedInteger(doc wirebson.AnyDocument, field string, required bool) (int64, error) {
	v, err := getOptionalParamAny(doc, field, nil)
	if err != nil {
		return 0, err
	}
	if v == nil && !required {
		return 0, nil
	}
	var n int64
	switch v := v.(type) {
	case int32:
		n = int64(v)
	case int64:
		n = v
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) || v != math.Trunc(v) || v >= float64(math.MaxInt64) || v < 0 {
			return 0, mongoerrors.New(mongoerrors.ErrBadValue, "Invalid capped collection size or count")
		}
		n = int64(v)
	default:
		return 0, mongoerrors.New(mongoerrors.ErrBadValue, "Capped collection size/count must be numeric")
	}
	if n < 0 || required && n == 0 {
		return 0, mongoerrors.New(mongoerrors.ErrBadValue, "Invalid capped collection size or count")
	}
	return n, nil
}

func ensureCappedSchema(ctx context.Context, conn *pgx.Conn) error {
	return pgx.BeginTxFunc(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout='5s'; SET LOCAL lock_timeout='2s'"); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended('ferretdb_unifi.capped.v1',0))"); err != nil {
			return err
		}
		var version string
		if err := tx.QueryRow(ctx, "SELECT extversion FROM pg_extension WHERE extname='documentdb'").Scan(&version); err != nil {
			return err
		}
		if version != "0.107-0" {
			return mongoerrors.New(mongoerrors.ErrNotImplemented, "Capped compatibility requires the validated DocumentDB version")
		}
		var exists, marker bool
		if err := tx.QueryRow(ctx, "SELECT to_regnamespace('ferretdb_unifi') IS NOT NULL,to_regclass('ferretdb_unifi.format_version') IS NOT NULL").Scan(&exists, &marker); err != nil {
			return err
		}
		if exists {
			if !marker {
				return mongoerrors.New(mongoerrors.ErrIllegalOperation, "Unrecognized compatibility schema")
			}
			var format int
			if err := tx.QueryRow(ctx, "SELECT version FROM ferretdb_unifi.format_version").Scan(&format); err != nil {
				return err
			}
			if format != 1 {
				return mongoerrors.New(mongoerrors.ErrNotImplemented, "Unsupported capped metadata format")
			}
			return nil
		}
		_, err := tx.Exec(ctx, cappedSchemaSQL)
		return err
	})
}

func (h *Handler) createCapped(ctx context.Context, db, name string, options cappedOptions, convert bool) error {
	return h.p.WithConnContext(ctx, func(conn *pgx.Conn) error {
		if err := ensureCappedSchema(ctx, conn); err != nil {
			return err
		}
		return pgx.BeginTxFunc(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", "ferretdb_unifi.capped."+db+"."+name); err != nil {
				return err
			}
			var id int64
			var view, sharded bool
			err := tx.QueryRow(ctx, "SELECT collection_id,view_definition IS NOT NULL,shard_key IS NOT NULL FROM documentdb_api_catalog.collections WHERE database_name=$1 AND collection_name=$2 FOR UPDATE", db, name).Scan(&id, &view, &sharded)
			if errors.Is(err, pgx.ErrNoRows) {
				if convert {
					return mongoerrors.New(mongoerrors.ErrNamespaceNotFound, "Collection does not exist")
				}
				created, e := documentdb_api.CreateCollection(ctx, tx.Conn(), h.L, db, name)
				if e != nil {
					return e
				}
				if !created {
					return mongoerrors.New(mongoerrors.ErrNamespaceExists, "Collection was concurrently created")
				}
				if err = tx.QueryRow(ctx, "SELECT collection_id FROM documentdb_api_catalog.collections WHERE database_name=$1 AND collection_name=$2", db, name).Scan(&id); err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if !convert {
				var same bool
				if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM ferretdb_unifi.capped_collections WHERE collection_id=$1 AND max_size=$2 AND max_documents=$3)", id, options.size, options.maximum).Scan(&same); err != nil {
					return err
				}
				if same {
					return nil
				}
				return mongoerrors.New(mongoerrors.ErrNamespaceExists, "Collection exists with different options")
			}
			if view || sharded {
				return mongoerrors.New(mongoerrors.ErrNotImplemented, "Capped compatibility supports plain collections")
			}
			table := pgx.Identifier{"documentdb_data", fmt.Sprintf("documents_%d", id)}.Sanitize()
			h.p.CloseNamespaceCursors(ctx, db, name)
			if _, err = tx.Exec(ctx, "LOCK TABLE "+table+" IN ACCESS EXCLUSIVE MODE"); err != nil {
				return err
			}
			var previousMaximum int64
			priorErr := tx.QueryRow(ctx, "SELECT max_documents FROM ferretdb_unifi.capped_collections WHERE collection_id=$1", id).Scan(&previousMaximum)
			alreadyCapped := priorErr == nil
			if priorErr != nil && !errors.Is(priorErr, pgx.ErrNoRows) {
				return priorErr
			}
			if alreadyCapped && convert {
				options.maximum = previousMaximum
			}
			if convert {
				spec, _ := wirebson.MustDocument("dropIndexes", name, "index", "*", "$db", db).Encode()
				if _, err = documentdb_api.DropIndexes(ctx, tx.Conn(), h.L, db, spec, nil); err != nil {
					return err
				}
				if _, err = tx.Exec(ctx, "UPDATE documentdb_api_catalog.collections SET collection_uuid=gen_random_uuid() WHERE collection_id=$1", id); err != nil {
					return err
				}
			}
			if alreadyCapped {
				if _, err = tx.Exec(ctx, "UPDATE ferretdb_unifi.capped_collections SET max_size=$2,max_documents=$3 WHERE collection_id=$1", id, options.size, options.maximum); err != nil {
					return err
				}
			} else {
				if _, err = tx.Exec(ctx, "INSERT INTO ferretdb_unifi.capped_collections(collection_id,max_size,max_documents) VALUES($1,$2,$3)", id, options.size, options.maximum); err != nil {
					return err
				}
				if _, err = tx.Exec(ctx, "INSERT INTO ferretdb_unifi.capped_order(collection_id,object_id,document_size) SELECT $1,object_id,octet_length(documentdb_core.bson_to_bytea(document)) FROM "+table+" ORDER BY ctid", id); err != nil {
					return err
				}
				if _, err = tx.Exec(ctx, "UPDATE ferretdb_unifi.capped_collections SET total_bytes=(SELECT COALESCE(sum(document_size),0) FROM ferretdb_unifi.capped_order WHERE collection_id=$1),total_documents=(SELECT count(*) FROM ferretdb_unifi.capped_order WHERE collection_id=$1) WHERE collection_id=$1", id); err != nil {
					return err
				}
			}
			if _, err = tx.Exec(ctx, "DROP TRIGGER IF EXISTS ferretdb_unifi_capped ON "+table); err != nil {
				return err
			}
			trigger := "CREATE TRIGGER ferretdb_unifi_capped AFTER INSERT OR UPDATE OR DELETE ON " + table + " FOR EACH ROW EXECUTE FUNCTION ferretdb_unifi.enforce_capped('" + fmt.Sprint(id) + "')"
			if _, err = tx.Exec(ctx, trigger); err != nil {
				return err
			}
			// Existing conversion rows are pruned in insertion order in this same transaction.
			for {
				var excess bool
				if err = tx.QueryRow(ctx, "SELECT total_documents>1 AND (total_bytes>max_size OR (max_documents>0 AND total_documents>max_documents)) FROM ferretdb_unifi.capped_collections WHERE collection_id=$1", id).Scan(&excess); err != nil {
					return err
				}
				if !excess {
					break
				}
				if _, err = tx.Exec(ctx, "DELETE FROM "+table+" WHERE object_id OPERATOR(documentdb_core.=) (SELECT object_id FROM ferretdb_unifi.capped_order WHERE collection_id=$1 ORDER BY position LIMIT 1)", id); err != nil {
					return err
				}
			}
			return nil
		})
	})
}

func (h *Handler) msgConvertToCapped(ctx context.Context, req *middleware.Request) (*middleware.Response, error) {
	if os.Getenv("FERRETDB_UNIFI_COMPAT") != "true" {
		return nil, mongoerrors.New(mongoerrors.ErrCommandNotFound, "no such command: convertToCapped")
	}
	doc := req.Document()
	db, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}
	name, err := getRequiredParam[string](doc, "convertToCapped")
	if err != nil {
		return nil, err
	}
	size, err := cappedInteger(doc, "size", true)
	if err != nil {
		return nil, err
	}
	if err = h.createCapped(ctx, db, name, cappedOptions{size: size}, true); err != nil {
		return nil, err
	}
	return middleware.ResponseDoc(req, wirebson.MustDocument("ok", float64(1)))
}

func (h *Handler) cappedMetadata(ctx context.Context, conn *pgx.Conn, db string) (map[string]cappedOptions, error) {
	result := map[string]cappedOptions{}
	var exists bool
	if err := conn.QueryRow(ctx, "SELECT to_regclass('ferretdb_unifi.capped_collections') IS NOT NULL").Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return result, nil
	}
	rows, err := conn.Query(ctx, "SELECT c.collection_name,m.max_size,m.max_documents FROM ferretdb_unifi.capped_collections m JOIN documentdb_api_catalog.collections c USING(collection_id) WHERE c.database_name=$1", db)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var options cappedOptions
		if err = rows.Scan(&name, &options.size, &options.maximum); err != nil {
			return nil, err
		}
		result[name] = options
	}
	return result, rows.Err()
}

func (h *Handler) cappedListPage(ctx context.Context, db string, page wirebson.RawDocument) (wirebson.RawDocument, error) {
	if os.Getenv("FERRETDB_UNIFI_COMPAT") != "true" {
		return page, nil
	}
	doc, err := page.Decode()
	if err != nil {
		return nil, err
	}
	doc = doc.Copy()
	value, ok := doc.Get("cursor").(wirebson.AnyDocument)
	if !ok {
		return page, nil
	}
	cursor, err := value.Decode()
	if err != nil {
		return nil, err
	}
	cursor = cursor.Copy()
	ns, _ := cursor.Get("ns").(string)
	if !strings.HasSuffix(strings.ToLower(ns), ".$cmd.listcollections") {
		return page, nil
	}
	var metadata map[string]cappedOptions
	if err = h.p.WithConnContext(ctx, func(conn *pgx.Conn) error { var e error; metadata, e = h.cappedMetadata(ctx, conn, db); return e }); err != nil {
		return nil, err
	}
	if len(metadata) == 0 {
		return page, nil
	}
	for _, key := range []string{"firstBatch", "nextBatch"} {
		array, ok := cursor.Get(key).(wirebson.AnyArray)
		if !ok {
			continue
		}
		batch, err := array.Decode()
		if err != nil {
			return nil, err
		}
		values := []any{}
		for entry := range batch.Values() {
			item, err := entry.(wirebson.AnyDocument).Decode()
			if err != nil {
				return nil, err
			}
			item = item.Copy()
			name, _ := item.Get("name").(string)
			if options, ok := metadata[name]; ok {
				opts := wirebson.MustDocument("capped", true, "size", cappedNumber(options.size))
				if options.maximum > 0 {
					if err = opts.Add("max", cappedNumber(options.maximum)); err != nil {
						return nil, err
					}
				}
				item.Remove("options")
				if err = item.Add("options", opts); err != nil {
					return nil, err
				}
			}
			values = append(values, item)
		}
		updated, err := wirebson.NewArray(values...)
		if err != nil {
			return nil, err
		}
		if err = cursor.Replace(key, updated); err != nil {
			return nil, err
		}
	}
	if err = doc.Replace("cursor", cursor); err != nil {
		return nil, err
	}
	return doc.Encode()
}
