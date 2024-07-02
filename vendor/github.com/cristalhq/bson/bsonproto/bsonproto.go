// Package bsonproto provides primitives for encoding and decoding of BSON.
package bsonproto

import (
	"errors"
	"fmt"
	"time"
)

// ScalarType represents a BSON scalar type.
//
// CString is not included as it is not a real BSON type.
type ScalarType interface {
	float64 | string | Binary | ObjectID | bool | time.Time | NullType | Regex | int32 | Timestamp | int64 | Decimal128
}

// Size returns a size of the encoding of value v in bytes.
func Size[T ScalarType](v T) int {
	return SizeAny(v)
}

// SizeAny returns a size of the encoding of value v in bytes.
//
// It panics if v is not a [ScalarType] (including CString).
func SizeAny(v any) int {
	switch v := v.(type) {
	case float64:
		return SizeFloat64
	case string:
		return SizeString(v)
	case Binary:
		return SizeBinary(v)
	case ObjectID:
		return SizeObjectID
	case bool:
		return SizeBool
	case time.Time:
		return SizeTime
	case NullType:
		return 0
	case Regex:
		return SizeRegex(v)
	case int32:
		return SizeInt32
	case Timestamp:
		return SizeTimestamp
	case int64:
		return SizeInt64
	case Decimal128:
		return SizeDecimal128
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}

// Encode encodes value v into b.
//
// b must be at least Size(v) bytes long; otherwise, Encode will panic.
// Only b[0:Size(v)] bytes are modified.
func Encode[T ScalarType](b []byte, v T) {
	EncodeAny(b, v)
}

// EncodeAny encodes value v into b.
//
// b must be at least Size(v) bytes long; otherwise, EncodeAny will panic.
// Only b[0:Size(v)] bytes are modified.
//
// It panics if v is not a [ScalarType] (including CString).
func EncodeAny(b []byte, v any) {
	switch v := v.(type) {
	case float64:
		EncodeFloat64(b, v)
	case string:
		EncodeString(b, v)
	case Binary:
		EncodeBinary(b, v)
	case ObjectID:
		EncodeObjectID(b, v)
	case bool:
		EncodeBool(b, v)
	case time.Time:
		EncodeTime(b, v)
	case NullType:
		// nothing
	case Regex:
		EncodeRegex(b, v)
	case int32:
		EncodeInt32(b, v)
	case Timestamp:
		EncodeTimestamp(b, v)
	case int64:
		EncodeInt64(b, v)
	case Decimal128:
		EncodeDecimal128(b, v)
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}

// Decode decodes value from b into v.
//
// If there is not enough bytes, Decode will return a wrapped [ErrDecodeShortInput].
// If the input is otherwise invalid, a wrapped [ErrDecodeInvalidInput] is returned.
func Decode[T ScalarType](b []byte, v *T) error {
	return DecodeAny(b, v)
}

// DecodeAny decodes value from b into v.
//
// If there is not enough bytes, DecodeAny will return a wrapped [ErrDecodeShortInput].
// If the input is otherwise invalid, a wrapped [ErrDecodeInvalidInput] is returned.
//
// It panics if v is not a pointer to [ScalarType] (including CString).
func DecodeAny(b []byte, v any) error {
	var err error
	switch v := v.(type) {
	case *float64:
		*v, err = DecodeFloat64(b)
	case *string:
		*v, err = DecodeString(b)
	case *Binary:
		*v, err = DecodeBinary(b)
	case *ObjectID:
		*v, err = DecodeObjectID(b)
	case *bool:
		*v, err = DecodeBool(b)
	case *time.Time:
		*v, err = DecodeTime(b)
	case *NullType:
		// nothing
	case *Regex:
		*v, err = DecodeRegex(b)
	case *int32:
		*v, err = DecodeInt32(b)
	case *Timestamp:
		*v, err = DecodeTimestamp(b)
	case *int64:
		*v, err = DecodeInt64(b)
	case *Decimal128:
		*v, err = DecodeDecimal128(b)
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}

	return err
}

var (
	// ErrDecodeShortInput is returned wrapped by Decode functions if the input bytes slice is too short.
	ErrDecodeShortInput = errors.New("bsonproto: short input")

	// ErrDecodeInvalidInput is returned wrapped by Decode functions if the input bytes slice is invalid.
	ErrDecodeInvalidInput = errors.New("bsonproto: invalid input")
)
