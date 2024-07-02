package bsonproto

import (
	"bytes"
	"fmt"
)

// SizeCString returns a size of the encoding of v cstring in bytes.
func SizeCString(v string) int {
	return len(v) + 1
}

// EncodeCString encodes cstring value v into b.
//
// b must be at least len(v)+1 ([SizeCString]) bytes long; otherwise, EncodeString will panic.
// Only b[0:len(v)+1] bytes are modified.
func EncodeCString(b []byte, v string) {
	// ensure b length early
	b[len(v)] = 0

	copy(b, v)
}

// DecodeCString decodes cstring value from b.
//
// If there is not enough bytes, DecodeCString will return a wrapped [ErrDecodeShortInput].
// If the input is otherwise invalid, a wrapped [ErrDecodeInvalidInput] is returned.
func DecodeCString(b []byte) (string, error) {
	if len(b) < 1 {
		return "", fmt.Errorf("DecodeCString: expected at least 1 byte, got %d: %w", len(b), ErrDecodeShortInput)
	}

	i := bytes.IndexByte(b, 0)
	if i == -1 {
		return "", fmt.Errorf("DecodeCString: expected to find 0 byte: %w", ErrDecodeInvalidInput)
	}

	return string(b[:i]), nil
}
