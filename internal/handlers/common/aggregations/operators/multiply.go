package operators

import (
	"errors"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

type multiply struct {
	// operators are documents containing operator expressions i.e. `[{$sum: 1}]`
	operators []*types.Document
	values    []any
}

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

			switch elemExpr := v.(type) {
			case *types.Document:
				if IsOperator(elemExpr) {
					operator.operators = append(operator.operators, elemExpr)
				}
			case float64:
				operator.values = append(operator.values, elemExpr)
			case string:
				ex, err := aggregations.NewExpression(elemExpr)

				var exErr *aggregations.ExpressionError
				if errors.As(err, &exErr) && exErr.Code() == aggregations.ErrNotExpression {
					break
				}

				if err != nil {
					return nil, err
				}

				operator.values = append(operator.values, ex)
			case int32, int64:
				operator.values = append(operator.values, elemExpr)
			}

		}
	}
}
