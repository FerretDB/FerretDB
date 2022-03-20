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
		case 'i':
			opts += "i"
		default:
			return nil, fmt.Errorf("types.Regex.Compile: unhandled regex option %v (%v)", o, r)
		}
	}

	expr := r.Pattern
	if opts != "" {
		expr = "(?" + opts + ")" + expr
	}

	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("types.Regex.Compile: %w", err)
	}

	return re, nil
}
