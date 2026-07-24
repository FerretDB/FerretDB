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

package common

import (
	"errors"
	"fmt"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Collation represents collation parameters for string comparison.
//
// A nil *Collation means the default binary ("simple") comparison.
type Collation struct {
	collator *collate.Collator
}

// NewCollation validates the given collation document and creates a Collation.
//
// It returns nil Collation (default binary comparison) for the nil document
// and for the "simple" locale.
func NewCollation(doc *types.Document) (*Collation, error) {
	if doc == nil {
		return nil, nil
	}

	if doc.Len() == 0 {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrBadValue,
			"BSON field 'collation' is an empty object",
			"collation",
		)
	}

	var locale string
	var strength int64 = 3

	numericOrdering := false

	iter := doc.Iterator()
	defer iter.Close()

	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return nil, lazyerrors.Error(err)
		}

		switch k {
		case "locale":
			s, ok := v.(string)
			if !ok {
				return nil, collationError("BSON field 'collation.locale' is the wrong type '%s', expected type 'string'",
					handlerparams.AliasFromType(v),
				)
			}

			locale = s

		case "strength":
			n, err := handlerparams.GetWholeNumberParam(v)
			if err != nil {
				return nil, collationError("BSON field 'collation.strength' is the wrong type '%s', "+
					"expected types '[long, int, decimal, double]'",
					handlerparams.AliasFromType(v),
				)
			}

			if n < 1 || n > 5 {
				return nil, collationError("BSON field 'collation.strength' value must be >= 1 and <= 5, actual value '%d'", n)
			}

			strength = n

		case "numericOrdering":
			b, ok := v.(bool)
			if !ok {
				return nil, collationError("BSON field 'collation.numericOrdering' is the wrong type '%s', "+
					"expected type 'bool'",
					handlerparams.AliasFromType(v),
				)
			}

			numericOrdering = b

		case "caseLevel":
			if b, ok := v.(bool); !ok || b {
				return nil, collationUnimplementedError(k, v)
			}

		case "backwards":
			if b, ok := v.(bool); !ok || b {
				return nil, collationUnimplementedError(k, v)
			}

		case "normalization":
			if _, ok := v.(bool); !ok {
				return nil, collationUnimplementedError(k, v)
			}

		case "caseFirst":
			if s, ok := v.(string); !ok || (s != "off") {
				return nil, collationUnimplementedError(k, v)
			}

		case "alternate":
			if s, ok := v.(string); !ok || (s != "non-ignorable") {
				return nil, collationUnimplementedError(k, v)
			}

		case "maxVariable":
			return nil, collationUnimplementedError(k, v)

		default:
			return nil, collationError("BSON field 'collation.%s' is an unknown field", k)
		}
	}

	if locale == "" {
		return nil, collationError("Missing expected field \"locale\"")
	}

	if locale == "simple" {
		// simple locale uses the default binary comparison
		return nil, nil
	}

	tag, err := language.Parse(locale)
	if err != nil {
		return nil, collationError("Field 'locale' is invalid in: { locale: \"%s\" }", locale)
	}

	var opts []collate.Option

	switch strength {
	case 1:
		opts = append(opts, collate.IgnoreCase, collate.IgnoreDiacritics)
	case 2:
		opts = append(opts, collate.IgnoreCase)
	default:
		// strengths 3, 4 and 5 use the default comparison
	}

	if numericOrdering {
		opts = append(opts, collate.Numeric)
	}

	return &Collation{collator: collate.New(tag, opts...)}, nil
}

// Compare compares two strings according to the collation.
// It returns a negative value if a < b, 0 if a == b, and a positive value if a > b.
func (c *Collation) Compare(a, b string) int {
	return c.collator.CompareString(a, b)
}

// collationError returns a formatted collation parsing error.
func collationError(format string, args ...any) error {
	return handlererrors.NewCommandErrorMsgWithArgument(
		handlererrors.ErrFailedToParse,
		fmt.Sprintf(format, args...),
		"collation",
	)
}

// collationUnimplementedError returns an error for valid but unsupported collation parameters.
func collationUnimplementedError(field string, value any) error {
	return handlererrors.NewCommandErrorMsgWithArgument(
		handlererrors.ErrNotImplemented,
		fmt.Sprintf("collation.%s %v is not supported yet", field, types.FormatAnyValue(value)),
		"collation",
	)
}
