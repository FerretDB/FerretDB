package handler

import (
	"strings"

	"github.com/FerretDB/wire/wirebson"
)

// unifiAggregateCommand preserves missing members of object-valued group keys.
// DocumentDB 0.107's direct group-key constructor replaces them with null,
// merging groups that MongoDB keeps separate. Ordinary expression evaluation
// retains missing values, so construct the key through arrayToObject instead.
func unifiAggregateCommand(raw wirebson.RawDocument) (wirebson.RawDocument, error) {
	doc, err := raw.DecodeDeep()
	if err != nil {
		return nil, err
	}
	changed, err := unifiAggregatePipeline(doc.Get("pipeline"))
	if err != nil {
		return nil, err
	}
	if !changed {
		return raw, nil
	}
	return doc.Encode()
}

func unifiAggregatePipeline(value any) (bool, error) {
	pipeline, ok := value.(*wirebson.Array)
	if !ok {
		return false, nil
	}
	changed := false
	for value := range pipeline.Values() {
		stage, ok := value.(*wirebson.Document)
		if !ok {
			continue
		}
		if group, ok := stage.Get("$group").(*wirebson.Document); ok {
			// MongoDB's one-member key optimization itself coalesces missing
			// and null. Preserve that native behavior for one-member objects.
			if key, ok := group.Get("_id").(*wirebson.Document); ok && len(key.FieldNames()) > 1 {
				object := true
				for _, name := range key.FieldNames() {
					if strings.HasPrefix(name, "$") {
						object = false
						break
					}
				}
				if object {
					pairs := make([]any, 0, len(key.FieldNames()))
					for _, name := range key.FieldNames() {
						pairs = append(pairs, wirebson.MustDocument("k", name, "v", key.Get(name)))
					}
					replacement := wirebson.MustDocument("$arrayToObject", wirebson.MustDocument("$filter", wirebson.MustDocument(
						"input", wirebson.MustArray(pairs...), "as", "ferretdbGroupPart",
						"cond", wirebson.MustDocument("$ne", wirebson.MustArray(wirebson.MustDocument("$type", "$$ferretdbGroupPart.v"), "missing")),
					)))
					if err := group.Replace("_id", replacement); err != nil {
						return false, err
					}
					changed = true
				}
			}
		}
		// Only descend into grammar-defined pipelines. In particular, a $literal
		// document containing "$group" is application data and remains untouched.
		for _, operator := range []string{"$lookup", "$unionWith", "$merge"} {
			if specification, ok := stage.Get(operator).(*wirebson.Document); ok {
				field := "pipeline"
				if operator == "$merge" {
					field = "whenMatched"
				}
				nested, err := unifiAggregatePipeline(specification.Get(field))
				if err != nil {
					return false, err
				}
				changed = changed || nested
			}
		}
		if facets, ok := stage.Get("$facet").(*wirebson.Document); ok {
			for _, name := range facets.FieldNames() {
				nested, err := unifiAggregatePipeline(facets.Get(name))
				if err != nil {
					return false, err
				}
				changed = changed || nested
			}
		}
	}
	return changed, nil
}
