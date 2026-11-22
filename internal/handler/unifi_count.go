package handler

import (
	"slices"

	"github.com/FerretDB/wire/wirebson"
)

// unifiCountCommand chooses DocumentDB's statement-snapshot count path instead
// of the relation-statistics shortcut used when query/skip/limit are absent.
// An explicit empty query has the same MongoDB predicate semantics and also
// works for documents without an _id in a view. Existing query values, including
// invalid types, remain untouched for native validation.
func unifiCountCommand(raw wirebson.RawDocument) (wirebson.RawDocument, error) {
	doc, err := raw.DecodeDeep()
	if err != nil {
		return nil, err
	}
	if slices.Contains(doc.FieldNames(), "query") {
		return raw, nil
	}
	if err := doc.Add("query", wirebson.MustDocument()); err != nil {
		return nil, err
	}
	return doc.Encode()
}
