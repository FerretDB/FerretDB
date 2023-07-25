package operators

import (
	"errors"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

type multiply struct {
	// operators are documents containing operator expressions i.e. `[{$sum: 1}]`
	operators []*types.Document
}

// newSum collects values that can be summed in `numbers`,
// finds nested operators if any, validates path expressions
// to populate `$sum` operator. It ignores values that are not summable.
func newMultiply(doc *types.Document) (Operator, error) {
	expr := must.NotFail(doc.Get("$multiply"))
	operator := new(multiply)

	switch expr := expr.(type) {
	case *types.Document:
		if IsOperator(expr) {
			operator.operators = []*types.Document{expr}
		}

	case *types.Array:
		iter := expr.Iterator()
		defer iter.Close()

		for {
			_, v, err := iter.Next()

			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			if err != nil {
				return nil, lazyerrors.Error(err)
			}

		}
	}
}
