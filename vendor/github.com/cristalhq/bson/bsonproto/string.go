package bsonproto

import (
	"encoding/binary"
	"fmt"
)

// SizeString returns a size of the encoding of v string in bytes.
func SizeString(v string) int {
	return len(v) + 5
}

// EncodeString encodes string value v into b.
//
// b must be at least len(v)+5 ([SizeString]) bytes long; otherwise, EncodeString will panic.
// Only b[0:len(v)+5] bytes are modified.
func EncodeString(b []byte, v string) {
	i := len(v) + 1

	// ensure b length early
	b[4+i-1] = 0

	binary.LittleEndian.PutUint32(b, uint32(i))
	copy(b[4:4+i-1], v)
}

// DecodeString decodes string value from b.
//
// If there is not enough bytes, DecodeString will return a wrapped [ErrDecodeShortInput].
// If the input is otherwise invalid, a wrapped [ErrDecodeInvalidInput] is returned.
func DecodeString(b []byte) (string, error) {
	if len(b) < 5 {
		return "", fmt.Errorf("DecodeString: expected at least 5 bytes, got %d: %w", len(b), ErrDecodeShortInput)
	}

	i := int(binary.LittleEndian.Uint32(b))
	if i < 1 {
		return "", fmt.Errorf("DecodeString: expected the prefix to be at least 1, got %d: %w", i, ErrDecodeInvalidInput)
	}
	if e := 4 + i; len(b) < e {
		return "", fmt.Errorf("DecodeString: expected at least %d bytes, got %d: %w", e, len(b), ErrDecodeShortInput)
	}
	if b[4+i-1] != 0 {
		return "", fmt.Errorf("DecodeString: expected the last byte to be 0: %w", ErrDecodeInvalidInput)
	}

	return string(b[4 : 4+i-1]), nil
}
