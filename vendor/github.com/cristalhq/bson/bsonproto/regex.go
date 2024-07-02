package bsonproto

import (
	"fmt"
)

// Regex represents BSON scalar type regular expression.
type Regex struct {
	Pattern string
	Options string
}

// SizeRegex returns a size of the encoding of v [Regex] in bytes.
func SizeRegex(v Regex) int {
	return len(v.Pattern) + len(v.Options) + 2
}

// EncodeRegex encodes [Regex] value v into b.
//
// b must be at least len(v.Pattern)+len(v.Options)+2 ([SizeRegex]) bytes long; otherwise, EncodeRegex will panic.
// Only b[0:len(v.Pattern)+len(v.Options)+2] bytes are modified.
func EncodeRegex(b []byte, v Regex) {
	// ensure b length early
	b[len(v.Pattern)+len(v.Options)+1] = 0

	copy(b, v.Pattern)
	b[len(v.Pattern)] = 0
	copy(b[len(v.Pattern)+1:], v.Options)
}

// DecodeRegex decodes [Regex] value from b.
//
// If there is not enough bytes, DecodeRegex will return a wrapped [ErrDecodeShortInput].
// If the input is otherwise invalid, a wrapped [ErrDecodeInvalidInput] is returned.
func DecodeRegex(b []byte) (Regex, error) {
	var res Regex

	if len(b) < 2 {
		return res, fmt.Errorf("DecodeRegex: expected at least 2 bytes, got %d: %w", len(b), ErrDecodeShortInput)
	}

	p, o := -1, -1
	for i, b := range b {
		if b == 0 {
			if p == -1 {
				p = i
			} else {
				o = i
				break
			}
		}
	}

	if o == -1 {
		return res, fmt.Errorf("DecodeRegex: expected two 0 bytes: %w", ErrDecodeShortInput)
	}

	res.Pattern = string(b[:p])
	res.Options = string(b[p+1 : o])

	return res, nil
}
