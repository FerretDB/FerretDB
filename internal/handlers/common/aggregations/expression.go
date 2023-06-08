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

package aggregations

import (
	"strings"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../../../bin/stringer -linecomment -type ExpressionErrorCode

// ExpressionErrorCode represents Expression error code.
type ExpressionErrorCode int

const (
	_ ExpressionErrorCode = iota

	// ErrNotExpression indicates that field is not an expression.
	ErrNotExpression

	// ErrInvalidExpression indicates that expression is invalid.
	ErrInvalidExpression

	// ErrEmptyFieldPath indicates that field path expression is empty.
	ErrEmptyFieldPath

	// ErrUndefinedVariable indicates that variable name is not defined.
	ErrUndefinedVariable

	// ErrEmptyVariable indicates that variable name is empty.
	ErrEmptyVariable
)

// ExpressionError describes an error that occurs while evaluating expression.
type ExpressionError struct {
	code ExpressionErrorCode
}

// newExpressionError creates a new ExpressionError.
func newExpressionError(code ExpressionErrorCode) error {
	return &ExpressionError{code: code}
}

// Error implements the error interface.
func (e *ExpressionError) Error() string {
	return e.code.String()
}

// Code returns the ExpressionError code.
func (e *ExpressionError) Code() ExpressionErrorCode {
	return e.code
}

// Expression is an expression constructed from field value.
type Expression struct {
	*ExpressionOpts
	path types.Path
}

// ExpressionOpts represents options used to modify behavior of Expression functions.
type ExpressionOpts struct {
	// TODO https://github.com/FerretDB/FerretDB/issues/2348

	// IgnoreArrays disables checking arrays for provided key.
	// So expression {"$v.foo"} won't match {"v":[{"foo":42}]}
	IgnoreArrays bool // defaults to false
}

// NewExpressionWithOpts creates a new instance by checking expression string.
// It can take additional opts that specify how expressions should be evaluated.
func NewExpressionWithOpts(expression string, opts *ExpressionOpts) (*Expression, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/2348
	var val string

	switch {
	case strings.HasPrefix(expression, "$$"):
		// `$$` indicates field is a variable.
		v := strings.TrimPrefix(expression, "$$")
		if v == "" {
			return nil, newExpressionError(ErrEmptyVariable)
		}

		if strings.HasPrefix(v, "$") {
			return nil, newExpressionError(ErrInvalidExpression)
		}

		// TODO https://github.com/FerretDB/FerretDB/issues/2275
		return nil, newExpressionError(ErrUndefinedVariable)
	case strings.HasPrefix(expression, "$"):
		// `$` indicates field is a path.
		val = strings.TrimPrefix(expression, "$")

		if val == "" {
			return nil, newExpressionError(ErrEmptyFieldPath)
		}
	default:
		return nil, newExpressionError(ErrNotExpression)
	}

	var err error

	path, err := types.NewPathFromString(val)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Expression{
		path:           path,
		ExpressionOpts: opts,
	}, nil
}

// NewExpression creates a new instance by checking expression string.
func NewExpression(expression string) (*Expression, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/2348
	return NewExpressionWithOpts(expression, new(ExpressionOpts))
}

// Evaluate gets the value at the path.
// It returns nil if the path does not exists.
func (e *Expression) Evaluate(doc *types.Document) any {
	path := e.path

	if path.Len() == 1 {
		val, err := doc.Get(path.String())
		if err != nil {
			return nil
		}

		return val
	}

	var isPrefixArray bool
	prefix := path.Prefix()

	if v, err := doc.Get(prefix); err == nil {
		if _, isArray := v.(*types.Array); isArray {
			isPrefixArray = true
		}
	}

	vals := e.getPathValue(doc, path)

	if len(vals) == 0 {
		if isPrefixArray {
			// when the prefix is array, return empty array.
			return must.NotFail(types.NewArray())
		}

		return nil
	}

	if len(vals) == 1 && !isPrefixArray {
		// when the prefix is not array, return the value
		return vals[0]
	}

	// when the prefix is array, return an array of value.
	arr := types.MakeArray(len(vals))
	for _, v := range vals {
		arr.Append(v)
	}

	return arr
}

// GetExpressionSuffix returns suffix of pathExpression.
func (e *Expression) GetExpressionSuffix() string {
	return e.path.Suffix()
}

// getPathValue go through each key of the path iteratively to
// find values that exist at suffix.
// An array may return multiple values.
// At each key of the path, it checks:
//   - if the document has the key.
//   - if the array contains documents which have the key. (This check can
//     be disabled by setting ExpressionOpts.IgnoreArrays field).
//
// It is different from `common.getDocumentsAtSuffix`, it does not find array item by
// array dot notation `foo.0.bar`. It returns empty array [] because using index
// such as `0` does not match using expression path.
func (e *Expression) getPathValue(doc *types.Document, path types.Path) []any {
	// TODO https://github.com/FerretDB/FerretDB/issues/2348
	keys := path.Slice()
	vals := []any{doc}

	for _, key := range keys {
		// embeddedVals are the values found at current key.
		var embeddedVals []any

		for _, valAtKey := range vals {
			switch val := valAtKey.(type) {
			case *types.Document:
				embeddedVal, err := val.Get(key)
				if err != nil {
					continue
				}

				embeddedVals = append(embeddedVals, embeddedVal)
			case *types.Array:
				if e.IgnoreArrays {
					continue
				}
				// iterate elements to get documents that contain the key.
				for j := 0; j < val.Len(); j++ {
					elem := must.NotFail(val.Get(j))

					docElem, isDoc := elem.(*types.Document)
					if !isDoc {
						continue
					}

					embeddedVal, err := docElem.Get(key)
					if err != nil {
						continue
					}

					embeddedVals = append(embeddedVals, embeddedVal)
				}

			default:
				// not a document or array, do nothing
			}
		}

		vals = embeddedVals
	}

	return vals
}
