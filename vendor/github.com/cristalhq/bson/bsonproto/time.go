package bsonproto

import (
	"encoding/binary"
	"fmt"
	"time"
)

// SizeTime is a size of the encoding of [time.Time] in bytes.
const SizeTime = 8

// EncodeTime encodes [time.Time] value v into b.
//
// b must be at least 8 ([SizeTime]) byte long; otherwise, EncodeTime will panic.
// Only b[0:8] bytes are modified.
func EncodeTime(b []byte, v time.Time) {
	binary.LittleEndian.PutUint64(b, uint64(v.UnixMilli()))
}

// DecodeTime decodes [time.Time] value from b.
//
// If there is not enough bytes, DecodeTime will return a wrapped [ErrDecodeShortInput].
func DecodeTime(b []byte) (time.Time, error) {
	var res time.Time

	if len(b) < SizeTime {
		return res, fmt.Errorf("DecodeTime: expected at least %d bytes, got %d: %w", SizeTime, len(b), ErrDecodeShortInput)
	}

	res = time.UnixMilli(int64(binary.LittleEndian.Uint64(b))).UTC()

	return res, nil
}
