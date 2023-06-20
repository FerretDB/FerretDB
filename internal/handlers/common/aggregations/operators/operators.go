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
)

// newOperatorFunc is a type for a function that creates a standard aggregation operator.
//
// By standard aggregation operator we mean any operator that is not accumulator.
// While accumulators perform operations on multiple documents
// (for example `$count` can count documents in each `$group` group),
// standard operators perform operations on a single document.
type newOperatorFunc func(expression *types.Document) (Operator, error)

// Operator is a common interface for standard aggregation operators.
type Operator interface {
	// Process document and returns the result of applying operator.
	Process(in *types.Document) (any, error)
}

// NewOperator returns operator from provided document.
// The document should look like: `{<$operator>: <operator-value>}`.
func NewOperator(doc any) (Operator, error) {
	operatorDoc, ok := doc.(*types.Document)
	if !ok {
		return nil, newOperatorError(
			ErrWrongType,
			"Invalid type of operator field (expected document)",
		)
	}

	iter := operatorDoc.Iterator()
	defer iter.Close()

	var operatorExists bool

	for {
		k, _, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if strings.HasPrefix(k, "$") {
			operatorExists = true
			break
		}
	}

	switch {
	case !operatorExists:
		return nil, newOperatorError(
			ErrNoOperator,
			"No operator in document",
		)

	case operatorDoc.Len() > 1:
		return nil, newOperatorError(
			ErrTooManyFields,
			"The operator field specifies more than one operator",
		)
	}

	operator := operatorDoc.Command()

	newOperator, supported := Operators[operator]
	_, unsupported := unsupportedOperators[operator]

	switch {
	case supported && unsupported:
		panic(fmt.Sprintf("operator %q is in both `operators` and `unsupportedOperators`", operator))
	case supported && !unsupported:
		return newOperator(operatorDoc)
	case !supported && unsupported:
		return nil, newOperatorError(
			ErrNotImplemented,
			fmt.Sprintf("The operator %s is not implemented yet", operator),
		)
	default:
		return nil, newOperatorError(
			ErrInvalidExpression,
			fmt.Sprintf("Unrecognized expression '%s'", operator),
		)
	}
}

// Operators maps all standard aggregation operators.
var Operators = map[string]newOperatorFunc{
	// sorted alphabetically
	// TODO https://github.com/FerretDB/FerretDB/issues/2680
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
	"$sum":              {},
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
