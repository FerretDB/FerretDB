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
	"fmt"
	"regexp"
	"regexp/syntax"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

var (
	// ErrOptionNotImplemented indicates unimplemented regex option.
	ErrOptionNotImplemented = fmt.Errorf("regex: option not implemented")

	// ErrMissingParen indicates missing parentheses in regex expression.
	ErrMissingParen = fmt.Errorf("Regular expression is invalid: missing )")

	// ErrMissingBracket indicates missing terminating ] for character class.
	ErrMissingBracket = fmt.Errorf("Regular expression is invalid: missing terminating ] for character class")

	// ErrInvalidEscape indicates invalid escape errors.
	ErrInvalidEscape = fmt.Errorf("Regular expression is invalid: PCRE does not support \\L, \\l, \\N{name}, \\U, or \\u")

	// ErrMissingTerminator indicates syntax error in subpattern name (missing terminator).
	ErrMissingTerminator = fmt.Errorf("Regular expression is invalid: syntax error in subpattern name (missing terminator)")

	// ErrUnmatchedParentheses indicates unmatched parentheses.
	ErrUnmatchedParentheses = fmt.Errorf("Regular expression is invalid: unmatched parentheses")

	// ErrTrailingBackslash indicates \\ at end of the pattern.
	ErrTrailingBackslash = fmt.Errorf("Regular expression is invalid: \\ at end of pattern")

	// ErrNothingToRepeat indicates invalid regex: nothing to repeat.
	ErrNothingToRepeat = fmt.Errorf("Regular expression is invalid: nothing to repeat")

	// ErrInvalidClassRange indicates that range out of order in character class.
	ErrInvalidClassRange = fmt.Errorf("Regular expression is invalid: range out of order in character class")

	// ErrUnsupportedPerlOp indicates unrecognized character after the grouping sequence start.
	ErrUnsupportedPerlOp = fmt.Errorf("Regular expression is invalid: unrecognized character after (? or (?-")

	// ErrInvalidRepeatSize indicates that the regular expression is too large.
	ErrInvalidRepeatSize = fmt.Errorf("Regular expression is invalid: regular expression is too large")
)

// Regex represents BSON type Regex.
type Regex struct {
	Pattern string
	Options string
}

// Compile returns Go Regexp object.
func (r Regex) Compile() (*regexp.Regexp, error) {
	var opts string
	for _, o := range r.Options {
		switch o {
		case 'i', 'm', 's':
			opts += string(o)
		case 'x':
			// TODO: https://github.com/FerretDB/FerretDB/issues/592
			return nil, ErrOptionNotImplemented
		default:
			continue
		}
	}

	expr := r.Pattern
	if opts != "" {
		expr = "(?" + opts + ")" + expr
	}

	re, err := regexp.Compile(expr)
	if err == nil {
		return re, nil
	}

	if err, ok := err.(*syntax.Error); ok {
		//nolint:exhaustive // we don't need to handle all possible errors there
		switch err.Code {
		case syntax.ErrInvalidCharRange:
			return nil, ErrInvalidClassRange
		case syntax.ErrInvalidEscape:
			return nil, ErrInvalidEscape
		case syntax.ErrInvalidNamedCapture:
			return nil, ErrMissingTerminator
		case syntax.ErrInvalidPerlOp:
			return nil, ErrUnsupportedPerlOp
		case syntax.ErrInvalidRepeatOp:
			return nil, ErrNothingToRepeat
		case syntax.ErrInvalidRepeatSize:
			return nil, ErrInvalidRepeatSize
		case syntax.ErrMissingBracket:
			return nil, ErrMissingBracket
		case syntax.ErrMissingParen:
			return nil, ErrMissingParen
		case syntax.ErrMissingRepeatArgument:
			return nil, ErrNothingToRepeat
		case syntax.ErrTrailingBackslash:
			return nil, ErrTrailingBackslash
		case syntax.ErrUnexpectedParen:
			return nil, ErrUnmatchedParentheses
		default:
			return nil, lazyerrors.Error(err)
		}
	}

	return nil, lazyerrors.Error(err)
}
