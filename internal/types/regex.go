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
)

var (
	ErrOptionNotImplemented = fmt.Errorf("regex: option not implemented")
	ErrMissingParen         = fmt.Errorf("Regular expression is invalid: missing )")
	ErrMissingBracket       = fmt.Errorf("Regular expression is invalid: missing terminating ] for character class")
	ErrInvalidEscape        = fmt.Errorf("Regular expression is invalid: PCRE does not support \\L, \\l, \\N{name}, \\U, or \\u")
	ErrMissingTerminator    = fmt.Errorf("Regular expression is invalid: syntax error in subpattern name (missing terminator)")
	ErrUnmatchedParentheses = fmt.Errorf("Regular expression is invalid: unmatched parentheses")
	ErrTrailingBackslash    = fmt.Errorf("Regular expression is invalid: \\ at end of pattern")
	ErrNothingToRepeat      = fmt.Errorf("Regular expression is invalid: nothing to repeat")
	ErrInvalidClassRange    = fmt.Errorf("Regular expression is invalid: range out of order in character class")
	ErrUnsupportedPerlOp    = fmt.Errorf("Regular expression is invalid: unrecognized character after (? or (?-")
	ErrInvalidRepeatSize    = fmt.Errorf("Regular expression is invalid: regular expression is too large")
)

// Regex represents BSON type Regex.
type Regex struct {
	Pattern string
	Options string
}

// Compile returns Go Regexp object.
func (r Regex) Compile() (*regexp.Regexp, error) {
	expr := r.Pattern

	var opts string
	for _, o := range r.Options {
		switch o {
		case 'i', 'm', 's':
			opts += string(o)
		case 'x':
			expr = freeSpacingParse(expr)
		default:
			continue
		}
	}

	if opts != "" {
		expr = "(?" + opts + ")" + expr
	}

	re, err := regexp.Compile(expr)
	if err == nil {
		return re, nil
	}

	if err, ok := err.(*syntax.Error); ok {
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
		case syntax.ErrInternalError, syntax.ErrInvalidCharClass, syntax.ErrInvalidUTF8:
			return nil, fmt.Errorf("types.Regex.Compile: %w", err)
		}
	}
	return nil, fmt.Errorf("types.Regex.Compile: %w", err)
}

/*
TODO:

			o{1 0} 	=> o{10}
			o{10}	=> oooooooooo
			o\{1 0}	=> o{10}

			PROBLEM:
			if the function takes "o{1 0}" as an argument
			1) it will remove all whitespaces => o{10}
			2) it'll pass "o{10} to regexp
			3) the regexp return "oooooooooo" in place of "o{10}"

			SOLUTION: precede every "{" with an escape character of it
*/
// Returns true if expr is a valid ending of Quantifier
// TODO: Make this cleaner
func isQuantifier(expr string) bool {
	comma, numAfterComma := false, false
	for i, c := range expr {
		switch {
		case '0' <= c && c <= '9':
			if comma && !numAfterComma {
				numAfterComma = true
			}
			continue
		case c == ',':
			if comma { // {1,2,}
				return false
			}
			if i < 1 { // {,1}
				return false
			}
			comma = true
		case c == '}':
			if i < 1 { // {}
				return false
			}
			if comma != numAfterComma { // {1,}
				return false
			}
			return true
		default:
			return false
		}
	}
	return false // ""
}

func freeSpacingParse(expr string) string {
	commentBlock, backslash, bracket, curly := false, false, false, false
	outExpr := ""

	for i, c := range expr {
		switch {
		case curly:
			if c == '}' {
				curly = false
			}

		case bracket:
			if c == ']' {
				bracket = false
			}

		case !backslash && commentBlock:
			if c == '\n' {
				commentBlock = false
			}
			continue

		case !backslash:
			switch c {
			case '{':
				if !isQuantifier(expr[i+1:]) {
					outExpr += "\\"
				} else {
					curly = true
				}
			case '\\': // escape characters
				backslash = true
			case '[': // square brackets content shouldn't be modified
				bracket = true
			case '#': // commments
				commentBlock = true
				continue
			case ' ', '\n', '\t': // remove whitespaces
				continue
			}
		default:
			backslash = false
		}

		outExpr += string(c)
	}
	return outExpr
}
