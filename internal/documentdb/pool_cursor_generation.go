package documentdb

import (
	"context"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/cursor"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

// cursorGeneration uses the existing operation connection, without opening a
// transaction or reserving another backend. A changed UUID detects a rename,
// capped conversion, or drop/recreate performed by another gateway process.
func (p *Pool) cursorGeneration(ctx context.Context, conn *pgx.Conn, database, collection string) (string, error) {
	if !p.unifiCompat || collection == "" {
		return "", nil
	}
	var generation string
	err := conn.QueryRow(ctx, `SELECT COALESCE((SELECT collection_uuid::text FROM documentdb_api_catalog.collections WHERE database_name=$1 AND collection_name=$2), '')`, database, collection).Scan(&generation)
	return generation, err
}

func (p *Pool) checkCursorGeneration(ctx context.Context, conn *pgx.Conn, id int64) error {
	dependencies := p.r.Dependencies(id)
	if len(dependencies) != 0 {
		actual, err := readNamespaceGenerations(ctx, conn, dependencies)
		if err != nil {
			p.r.CloseCursor(ctx, id)
			return err
		}
		for i, dependency := range dependencies {
			if dependency.UUID != actual[i].UUID {
				p.r.CloseCursor(ctx, id)
				if dependency.ErrorCode != 0 && dependency.ErrorCode != 175 {
					return mongoerrors.New(mongoerrors.Code(dependency.ErrorCode), "Collection identity changed")
				}
				return queryPlanKilled()
			}
		}
	}
	database, collection, expected, tracked := p.r.Generation(id)
	if !tracked {
		return nil
	}
	actual, err := p.cursorGeneration(ctx, conn, database, collection)
	if err != nil {
		p.r.CloseCursor(ctx, id)
		return err
	}
	if actual == expected {
		return nil
	}
	p.r.CloseCursor(ctx, id)
	return queryPlanKilled()
}

func queryPlanKilled() *mongoerrors.Error {
	killed := mongoerrors.New(mongoerrors.Code(175), "Collection identity changed")
	killed.Name = "QueryPlanKilled"
	return killed
}

// lookupNamespaces visits pipeline grammar only, never literal BSON application data.
// Union stages have deferred child-cursor semantics, not eager lookup dependencies.
func lookupNamespaces(database string, value any, seen map[[2]string]bool, result *[]cursor.NamespaceGeneration) {
	collectLookupNamespaces(database, value, seen, nil, result)
}

func collectLookupNamespaces(database string, value any, seen map[[2]string]bool, known map[[2]string]string, result *[]cursor.NamespaceGeneration) {
	pipeline, ok := value.(*wirebson.Array)
	if !ok {
		return
	}
	for value := range pipeline.Values() {
		stage, ok := value.(*wirebson.Document)
		if !ok {
			continue
		}
		for _, operator := range []string{"$lookup", "$graphLookup"} {
			specification, ok := stage.Get(operator).(*wirebson.Document)
			if !ok {
				continue
			}
			db, collection := database, ""
			switch from := specification.Get("from").(type) {
			case string:
				collection = from
			case *wirebson.Document:
				if name, ok := from.Get("db").(string); ok {
					db = name
				}
				collection, _ = from.Get("coll").(string)
			}
			key := [2]string{db, collection}
			if collection != "" && known != nil && known[key] == "" {
				// A lookup against an initially absent collection has an EOF plan.
				// Later collection creation does not invalidate that captured plan.
				continue
			}
			if collection != "" && !seen[key] {
				seen[key] = true
				code := int32(175)
				if operator == "$graphLookup" {
					code = 26
				}
				*result = append(*result, cursor.NamespaceGeneration{Database: db, Collection: collection, UUID: known[key], ErrorCode: code})
			}
			collectLookupNamespaces(db, specification.Get("pipeline"), seen, known, result)
		}
		if facets, ok := stage.Get("$facet").(*wirebson.Document); ok {
			for _, name := range facets.FieldNames() {
				collectLookupNamespaces(database, facets.Get(name), seen, known, result)
			}
		}
	}
}

func readNamespaceGenerations(ctx context.Context, conn *pgx.Conn, dependencies []cursor.NamespaceGeneration) ([]cursor.NamespaceGeneration, error) {
	databases, collections := make([]string, len(dependencies)), make([]string, len(dependencies))
	for i, dependency := range dependencies {
		databases[i], collections[i] = dependency.Database, dependency.Collection
	}
	rows, err := conn.Query(ctx, `SELECT n.database_name, n.collection_name, COALESCE(c.collection_uuid::text, '')
		FROM unnest($1::text[], $2::text[]) WITH ORDINALITY AS n(database_name,collection_name,position)
		LEFT JOIN documentdb_api_catalog.collections c USING (database_name,collection_name) ORDER BY n.position`, databases, collections)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]cursor.NamespaceGeneration, 0, len(dependencies))
	for rows.Next() {
		var generation cursor.NamespaceGeneration
		if err := rows.Scan(&generation.Database, &generation.Collection, &generation.UUID); err != nil {
			return nil, err
		}
		result = append(result, generation)
	}
	return result, rows.Err()
}

func (p *Pool) aggregateDependencies(ctx context.Context, conn *pgx.Conn, db string, spec wirebson.RawDocument) ([]cursor.NamespaceGeneration, error) {
	if !p.unifiCompat {
		return nil, nil
	}
	document, err := spec.DecodeDeep()
	if err != nil {
		return nil, err
	}
	var dependencies []cursor.NamespaceGeneration
	lookupNamespaces(db, document.Get("pipeline"), make(map[[2]string]bool), &dependencies)
	if len(dependencies) == 0 {
		return nil, nil
	}
	observed, err := readNamespaceGenerations(ctx, conn, dependencies)
	if err != nil {
		return nil, err
	}
	known := make(map[[2]string]string, len(observed))
	for _, dependency := range observed {
		known[[2]string{dependency.Database, dependency.Collection}] = dependency.UUID
	}
	var active []cursor.NamespaceGeneration
	collectLookupNamespaces(db, document.Get("pipeline"), make(map[[2]string]bool), known, &active)
	return active, nil
}
