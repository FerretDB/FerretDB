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

package types

import (
	"strings"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../bin/stringer -linecomment -type ExpressionErrorCode

// ExpressionErrorCode represents FieldPath error code.
type ExpressionErrorCode int

const (
	_ ExpressionErrorCode = iota

	// ErrNotFieldPath indicates that field is not a path.
	ErrNotFieldPath

	// ErrEmptyFieldPath indicates that path is empty.
	ErrEmptyFieldPath

	// ErrInvalidFieldPath indicates that path is invalid.
	ErrInvalidFieldPath

	// ErrUndefinedVariable indicates that variable name is not defined.
	ErrUndefinedVariable

	// ErrEmptyVariable indicates that variable name is empty.
	ErrEmptyVariable
)

// FieldPathError describes an error that occurs getting path from field.
type FieldPathError struct {
	code ExpressionErrorCode
}

// newFieldPathError creates a new FieldPathError.
func newFieldPathError(code ExpressionErrorCode) error {
	return &FieldPathError{code: code}
}

// Error implements the error interface.
func (e *FieldPathError) Error() string {
	return e.code.String()
}

// Code returns the FieldPathError code.
func (e *FieldPathError) Code() ExpressionErrorCode {
	return e.code
}

// Expression is an expression constructed from field value.
type Expression interface {
	Evaluate(doc *Document) any
	GetPath() Path
}

// pathExpression is field path constructed from expression.
type pathExpression struct {
	path Path
	*ExpressionOpts
}

// ExpressionOpts represents options used to modify behavior of Expression functions.
type ExpressionOpts struct {
	IgnoreArrays bool // defaults to false
}

// NewExpressionWithOpts creates a new instance by checking expression string.
// It can take additional opts that specify how expressions should be evaluated.
func NewExpressionWithOpts(expression string, opts *ExpressionOpts) (Expression, error) {
	var val string

	switch {
	case strings.HasPrefix(expression, "$$"):
		// `$$` indicates field is a variable.
		v := strings.TrimPrefix(expression, "$$")
		if v == "" {
			return nil, newFieldPathError(ErrEmptyVariable)
		}

		if strings.HasPrefix(v, "$") {
			return nil, newFieldPathError(ErrInvalidFieldPath)
		}

		// TODO https://github.com/FerretDB/FerretDB/issues/2275
		return nil, newFieldPathError(ErrUndefinedVariable)
	case strings.HasPrefix(expression, "$"):
		// `$` indicates field is a path.
		val = strings.TrimPrefix(expression, "$")

		if val == "" {
			return nil, newFieldPathError(ErrEmptyFieldPath)
		}
	default:
		return nil, newFieldPathError(ErrNotFieldPath)
	}

	var err error

	path, err := NewPathFromString(val)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &pathExpression{
		path:           path,
		ExpressionOpts: opts,
	}, nil
}

// NewExpression creates a new instance by checking expression string.
func NewExpression(expression string) (Expression, error) {
	return NewExpressionWithOpts(expression, new(ExpressionOpts))
}

// Evaluate gets the value at the path.
func (p *pathExpression) Evaluate(doc *Document) any {
	path := p.path

	if path.Len() == 1 {
		val, err := doc.Get(path.String())
		if err != nil {
			// if the path does not exist, return nil.
			return Null
		}

		return val
	}

	var isPrefixArray bool
	prefix := path.Prefix()

	if v, err := doc.Get(prefix); err == nil {
		if _, isArray := v.(*Array); isArray {
			isPrefixArray = true
		}
	}

	vals := p.getExpressionPathValue(doc, path)

	if len(vals) == 0 {
		if isPrefixArray {
			// when the prefix is array, return empty array.
			return must.NotFail(NewArray())
		}

		return Null
	}

	if len(vals) == 1 && !isPrefixArray {
		// when the prefix is not array, return the value
		return vals[0]
	}

	// when the prefix is array, return an array of value.
	arr := MakeArray(len(vals))
	for _, v := range vals {
		arr.Append(v)
	}

	return arr
}

// GetPath returns path of expression.
func (p *pathExpression) GetPath() Path {
	return p.path
}

// getExpressionPathValue go through each key of the path iteratively to
// find values that exist at suffix.
// An array may return multiple values.
// At each key of the path, it checks:
// - if the document has the key.
// - if the array contains documents which have the key. (This check can
// be disabled by setting ExpressionOpts.IgnoreArrays field).
//
// It is different from `getDocumentsAtSuffix`, it does not find array item by
// array dot notation `foo.0.bar`. It returns empty array [] because using index
// such as `0` does not match using expression path.
func (p *pathExpression) getExpressionPathValue(doc *Document, path Path) []any {
	keys := path.Slice()
	vals := []any{doc}

	for _, key := range keys {
		// embeddedVals are the values found at current key.
		var embeddedVals []any

		for _, valAtKey := range vals {
			switch val := valAtKey.(type) {
			case *Document:
				embeddedVal, err := val.Get(key)
				if err != nil {
					continue
				}

				embeddedVals = append(embeddedVals, embeddedVal)
			case *Array:
				if p.IgnoreArrays {
					continue
				}
				// iterate elements to get documents that contain the key.
				for j := 0; j < val.Len(); j++ {
					elem := must.NotFail(val.Get(j))

					docElem, isDoc := elem.(*Document)
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
