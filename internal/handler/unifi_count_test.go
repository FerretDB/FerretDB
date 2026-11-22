package handler

import (
	"testing"

	"github.com/FerretDB/wire/wirebson"
	"github.com/stretchr/testify/require"
)

func TestUniFiCountSelectsExactPathWithoutChangingOptions(t *testing.T) {
	raw, err := wirebson.MustDocument("count", "synthetic", "skip", int32(1), "limit", int64(2), "maxTimeMS", int32(100), "hint", "synthetic_index", "$db", "synthetic-db").Encode()
	require.NoError(t, err)
	result, err := unifiCountCommand(raw)
	require.NoError(t, err)
	doc, err := result.DecodeDeep()
	require.NoError(t, err)
	require.Equal(t, "count", doc.Command())
	require.Equal(t, int32(1), doc.Get("skip"))
	require.Equal(t, int64(2), doc.Get("limit"))
	require.Equal(t, int32(100), doc.Get("maxTimeMS"))
	require.Equal(t, "synthetic_index", doc.Get("hint"))
	require.Equal(t, "synthetic-db", doc.Get("$db"))
	require.Empty(t, doc.Get("query").(*wirebson.Document).FieldNames())
	again, err := unifiCountCommand(result)
	require.NoError(t, err)
	require.Equal(t, result, again)
}

func TestUniFiCountPreservesExistingQueriesAndValidation(t *testing.T) {
	for _, query := range []any{wirebson.MustDocument(), wirebson.MustDocument("n", int32(1)), wirebson.Null, "invalid", int32(1), wirebson.MustArray()} {
		raw, err := wirebson.MustDocument("count", "synthetic", "query", query).Encode()
		require.NoError(t, err)
		result, err := unifiCountCommand(raw)
		require.NoError(t, err)
		require.Equal(t, raw, result)
	}
}
