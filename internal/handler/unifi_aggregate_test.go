package handler

import (
	"bytes"
	"testing"

	"github.com/FerretDB/wire/wirebson"
	"github.com/stretchr/testify/require"
)

func TestUniFiAggregateGroupKeys(t *testing.T) {
	objectGroup := func() *wirebson.Document {
		return wirebson.MustDocument("$group", wirebson.MustDocument("_id", wirebson.MustDocument("present", "$present", "optional", "$optional"), "n", wirebson.MustDocument("$sum", int32(1))))
	}
	for _, tc := range []struct {
		name     string
		pipeline *wirebson.Array
		changed  bool
	}{
		{"object", wirebson.MustArray(objectGroup()), true},
		{"scalar", wirebson.MustArray(wirebson.MustDocument("$group", wirebson.MustDocument("_id", "$present"))), false},
		{"singleMember", wirebson.MustArray(wirebson.MustDocument("$group", wirebson.MustDocument("_id", wirebson.MustDocument("optional", "$optional")))), false},
		{"emptyObject", wirebson.MustArray(wirebson.MustDocument("$group", wirebson.MustDocument("_id", wirebson.MustDocument()))), false},
		{"expression", wirebson.MustArray(wirebson.MustDocument("$group", wirebson.MustDocument("_id", wirebson.MustDocument("$literal", wirebson.MustDocument("optional", "$optional"))))), false},
		{"literalData", wirebson.MustArray(wirebson.MustDocument("$project", wirebson.MustDocument("value", wirebson.MustDocument("$literal", objectGroup())))), false},
		{"facet", wirebson.MustArray(wirebson.MustDocument("$facet", wirebson.MustDocument("nested", wirebson.MustArray(objectGroup())))), true},
		{"lookup", wirebson.MustArray(wirebson.MustDocument("$lookup", wirebson.MustDocument("from", "synthetic", "pipeline", wirebson.MustArray(objectGroup()), "as", "joined"))), true},
		{"union", wirebson.MustArray(wirebson.MustDocument("$unionWith", wirebson.MustDocument("coll", "synthetic", "pipeline", wirebson.MustArray(objectGroup())))), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := wirebson.MustDocument("aggregate", "synthetic", "pipeline", tc.pipeline, "cursor", wirebson.MustDocument(), "maxTimeMS", int32(100)).Encode()
			require.NoError(t, err)
			result, err := unifiAggregateCommand(raw)
			require.NoError(t, err)
			require.Equal(t, tc.changed, !bytes.Equal(raw, result))
			decoded, err := result.Decode()
			require.NoError(t, err)
			require.Equal(t, int32(100), decoded.Get("maxTimeMS"))
			if tc.changed {
				require.Contains(t, string(result), "ferretdbGroupPart")
			}
			// Rewriting an expression key again should have no effect.
			again, err := unifiAggregateCommand(result)
			require.NoError(t, err)
			require.Equal(t, result, again)
		})
	}
}
