package jsonb1

import (
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
)

func projection(projection types.Document, p *pg.Placeholder) (sql string, args []any, err error) {
	projectionMap := projection.Map()
	if len(projectionMap) == 0 {
		sql = "_jsonb"
		return
	}

	ks := ""
	for i, k := range projection.Keys() {
		if i != 0 {
			ks += ", "
		}
		ks += p.Next()
		args = append(args, k)
	}
	sql = "json_build_object('$k', array[" + ks + "],"
	for i, k := range projection.Keys() {
		if i != 0 {
			sql += ", "
		}
		sql += p.Next() + "::text, _jsonb->" + p.Next()
		args = append(args, k, k)
	}
	sql += ")"

	return
}
