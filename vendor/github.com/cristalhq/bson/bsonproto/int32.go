package bsonproto

import (
	"encoding/binary"
	"fmt"
)

// SizeInt32 is a size of the encoding of int32 in bytes.
const SizeInt32 = 4

// EncodeInt32 encodes int32 value v into b.
//
// b must be at least 4 ([SizeInt32]) bytes long; otherwise, EncodeInt32 will panic.
// Only b[0:4] bytes are modified.
func EncodeInt32(b []byte, v int32) {
	binary.LittleEndian.PutUint32(b, uint32(v))
}

// DecodeInt32 decodes int32 value from b.
//
// If there is not enough bytes, DecodeInt32 will return a wrapped [ErrDecodeShortInput].
func DecodeInt32(b []byte) (int32, error) {
	if len(b) < SizeInt32 {
		return 0, fmt.Errorf("DecodeInt32: expected at least %d bytes, got %d: %w", SizeInt32, len(b), ErrDecodeShortInput)
	}

	return int32(binary.LittleEndian.Uint32(b)), nil
}
