package documentdb

import (
	"reflect"
	"testing"

	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/cursor"
)

func TestUniFiCursorDependencies(t *testing.T) {
	lookup := func(from any, pipeline ...any) *wirebson.Document {
		return wirebson.MustDocument("$lookup", wirebson.MustDocument("from", from, "pipeline", wirebson.MustArray(pipeline...)))
	}
	pipeline := wirebson.MustArray(
		lookup("joined", lookup(wirebson.MustDocument("db", "other", "coll", "nested"))),
		lookup("joined"),
		wirebson.MustDocument("$facet", wirebson.MustDocument("branch", wirebson.MustArray(
			wirebson.MustDocument("$graphLookup", wirebson.MustDocument("from", "graph")),
		))),
		wirebson.MustDocument("$project", wirebson.MustDocument("literal", wirebson.MustDocument("$literal", lookup("application-data")))),
		wirebson.MustDocument("$unionWith", wirebson.MustDocument("coll", "deferred", "pipeline", wirebson.MustArray(lookup("deferred-child")))),
	)
	var actual []cursor.NamespaceGeneration
	lookupNamespaces("source", pipeline, make(map[[2]string]bool), &actual)
	expected := []cursor.NamespaceGeneration{
		{Database: "source", Collection: "joined", ErrorCode: 175},
		{Database: "other", Collection: "nested", ErrorCode: 175},
		{Database: "source", Collection: "graph", ErrorCode: 26},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("pipeline dependency mismatch: got %#v, want %#v", actual, expected)
	}
	known := map[[2]string]string{
		{"source", "joined"}: "",
		{"other", "nested"}:  "existing-but-unreached",
		{"source", "graph"}:  "graph-uuid",
	}
	actual = nil
	collectLookupNamespaces("source", pipeline, make(map[[2]string]bool), known, &actual)
	expected = []cursor.NamespaceGeneration{{Database: "source", Collection: "graph", UUID: "graph-uuid", ErrorCode: 26}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("initially absent lookup retained a live dependency: %#v", actual)
	}
}
