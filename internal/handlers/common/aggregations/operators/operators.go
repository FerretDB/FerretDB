// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package operators provides aggregation operators.
// Operators are used in aggregation stages to filter and model data.
// This package contains all operators apart from the accumulation operators,
// which are stored and described in accumulators package.
//
// Accumulators that can be used outside of accumulation with different behaviour (like `$sum`),
// should be stored in both operators and accumulators packages.
package operators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// newOperatorFunc is a type for a function that creates a standard aggregation operator.
//
// By standard aggregation operator we mean any operator that is not accumulator.
// While accumulators perform operations on multiple documents
// (for example `$count` can count documents in each `$group` group),
// standard operators perform operations on a single document.
// It takes the arguments extracted from the document, and not the
// whole array/document.
type newOperatorFunc func(args ...any) (Operator, error)

// Operator is a common interface for standard aggregation operators.
type Operator interface {
	// Process document and returns the result of applying operator.
	Process(in *types.Document) (any, error)
}

// IsOperator returns true if provided document should be
// treated as operator document.
func IsOperator(doc *types.Document) bool {
	iter := doc.Iterator()
	defer iter.Close()

	for {
		key, _, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return false
		}

		if strings.HasPrefix(key, "$") {
			return true
		}
	}

	return false
}

// NewOperator returns operator from provided document.
// The document should look like: `{<$operator>: <operator-value>}`.
//
// Before calling NewOperator on document it's recommended to validate
// document before by using IsOperator on it.
func NewOperator(doc *types.Document) (Operator, error) {
	if doc.Len() == 0 {
		return nil, lazyerrors.New(
			"The operator field is empty",
		)
	}

	if doc.Len() > 1 {
		return nil, newOperatorError(
			ErrTooManyFields,
			doc.Command(),
			"The operator field specifies more than one operator",
		)
	}

	operator := doc.Command()

	newOperator, supported := Operators[operator]
	_, unsupported := unsupportedOperators[operator]

	expr := must.NotFail(doc.Get(operator))

	var args []any

	if arr, ok := expr.(*types.Array); ok {
		iter := arr.Iterator()
		defer iter.Close()

		for {
			_, v, err := iter.Next()

			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			args = append(args, v)
		}
	} else {
		args = append(args, expr)
	}

	switch {
	case supported:
		return newOperator(args...)
	case unsupported:
		return nil, newOperatorError(
			ErrNotImplemented,
			operator,
			fmt.Sprintf("The operator %s is not implemented yet", operator),
		)
	default:
		return nil, newOperatorError(
			ErrInvalidExpression,
			operator,
			fmt.Sprintf("Unrecognized expression '%s'", operator),
		)
	}
}

// Operators maps all standard aggregation operators.
var Operators = map[string]newOperatorFunc{
	// sorted alphabetically
	"$sum":  newSum,
	"$type": newType,
	// please keep sorted alphabetically
}

// unsupportedOperators maps all unsupported yet operators.
var unsupportedOperators = map[string]struct{}{
	// sorted alphabetically
	"$abs":              {},
	"$acos":             {},
	"$acosh":            {},
	"$add":              {},
	"$allElementsTrue":  {},
	"$and":              {},
	"$anyElementTrue":   {},
	"$arrayElemAt":      {},
	"$arrayToObject":    {},
	"$asin":             {},
	"$asinh":            {},
	"$atan":             {},
	"$atan2":            {},
	"$atanh":            {},
	"$avg":              {},
	"$binarySize":       {},
	"$bsonSize":         {},
	"$ceil":             {},
	"$cmp":              {},
	"$concat":           {},
	"$concatArrays":     {},
	"$cond":             {},
	"$convert":          {},
	"$cos":              {},
	"$cosh":             {},
	"$covariancePop":    {},
	"$covarianceSamp":   {},
	"$dateAdd":          {},
	"$dateDiff":         {},
	"$dateFromParts":    {},
	"$dateSubtract":     {},
	"$dateTrunc":        {},
	"$dateToParts":      {},
	"$dateFromString":   {},
	"$dateToString":     {},
	"$dayOfMonth":       {},
	"$dayOfWeek":        {},
	"$dayOfYear":        {},
	"$degreesToRadians": {},
	"$denseRank":        {},
	"$derivative":       {},
	"$divide":           {},
	"$documentNumber":   {},
	"$eq":               {},
	"$exp":              {},
	"$expMovingAvg":     {},
	"$filter":           {},
	"$floor":            {},
	"$function":         {},
	"$getField":         {},
	"$gt":               {},
	"$gte":              {},
	"$hour":             {},
	"$ifNull":           {},
	"$in":               {},
	"$indexOfArray":     {},
	"$indexOfBytes":     {},
	"$indexOfCP":        {},
	"$integral":         {},
	"$isArray":          {},
	"$isNumber":         {},
	"$isoDayOfWeek":     {},
	"$isoWeek":          {},
	"$isoWeekYear":      {},
	"$let":              {},
	"$linearFill":       {},
	"$literal":          {},
	"$ln":               {},
	"$locf":             {},
	"$log":              {},
	"$log10":            {},
	"$lt":               {},
	"$lte":              {},
	"$ltrim":            {},
	"$map":              {},
	"$max":              {},
	"$meta":             {},
	"$min":              {},
	"$minN":             {},
	"$millisecond":      {},
	"$minute":           {},
	"$mod":              {},
	"$month":            {},
	"$multiply":         {},
	"$ne":               {},
	"$not":              {},
	"$objectToArray":    {},
	"$or":               {},
	"$pow":              {},
	"$radiansToDegrees": {},
	"$rand":             {},
	"$range":            {},
	"$rank":             {},
	"$reduce":           {},
	"$regexFind":        {},
	"$regexFindAll":     {},
	"$regexMatch":       {},
	"$replaceOne":       {},
	"$replaceAll":       {},
	"$reverseArray":     {},
	"$round":            {},
	"$rtrim":            {},
	"$sampleRate":       {},
	"$second":           {},
	"$setDifference":    {},
	"$setEquals":        {},
	"$setField":         {},
	"$setIntersection":  {},
	"$setIsSubset":      {},
	"$setUnion":         {},
	"$shift":            {},
	"$size":             {},
	"$sin":              {},
	"$sinh":             {},
	"$slice":            {},
	"$sortArray":        {},
	"$split":            {},
	"$sqrt":             {},
	"$stdDevPop":        {},
	"$stdDevSamp":       {},
	"$strcasecmp":       {},
	"$strLenBytes":      {},
	"$strLenCP":         {},
	"$substr":           {},
	"$substrBytes":      {},
	"$substrCP":         {},
	"$subtract":         {},
	"$switch":           {},
	"$tan":              {},
	"$tanh":             {},
	"$toBool":           {},
	"$toDate":           {},
	"$toDecimal":        {},
	"$toDouble":         {},
	"$toInt":            {},
	"$toLong":           {},
	"$toObjectId":       {},
	"$toString":         {},
	"$toLower":          {},
	"$toUpper":          {},
	"$trim":             {},
	"$trunc":            {},
	"$tsIncrement":      {},
	"$tsSecond":         {},
	"$unsetField":       {},
	"$week":             {},
	"$year":             {},
	"$zip":              {},
	// please keep sorted alphabetically
}
