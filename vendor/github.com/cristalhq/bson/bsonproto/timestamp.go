package bsonproto

import (
	"encoding/binary"
	"fmt"
)

// Timestamp represents BSON scalar type timestamp.
type Timestamp uint64

// SizeTimestamp is a size of the encoding of [Timestamp] in bytes.
const SizeTimestamp = 8

// EncodeTimestamp encodes [Timestamp] value v into b.
//
// b must be at least 8 ([SizeTimestamp]) bytes long; otherwise, EncodeTimestamp will panic.
// Only b[0:8] bytes are modified.
func EncodeTimestamp(b []byte, v Timestamp) {
	binary.LittleEndian.PutUint64(b, uint64(v))
}

// DecodeTimestamp decodes [Timestamp] value from b.
//
// If there is not enough bytes, DecodeTimestamp will return a wrapped [ErrDecodeShortInput].
func DecodeTimestamp(b []byte) (Timestamp, error) {
	if len(b) < SizeTimestamp {
		return 0, fmt.Errorf("DecodeTimestamp: expected at least %d bytes, got %d: %w", SizeTimestamp, len(b), ErrDecodeShortInput)
	}

	return Timestamp(binary.LittleEndian.Uint64(b)), nil
}
