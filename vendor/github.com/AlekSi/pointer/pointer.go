// Package pointer provides helpers to convert between pointers and values of built-in (and, with generics, of any) types.
package pointer // import "github.com/AlekSi/pointer"

import (
	"time"
)

/*
Order as in spec:
	bool byte complex64 complex128 error float32 float64
	int int8 int16 int32 int64 rune string
	uint uint8 uint16 uint32 uint64 uintptr
	time.Duration time.Time
*/

// ToBool returns a pointer to the passed bool value.
func ToBool(b bool) *bool {
	return &b
}

// ToByte returns a pointer to the passed byte value.
func ToByte(b byte) *byte {
	return &b
}

// ToComplex64 returns a pointer to the passed complex64 value.
func ToComplex64(c complex64) *complex64 {
	return &c
}

// ToComplex128 returns a pointer to the passed complex128 value.
func ToComplex128(c complex128) *complex128 {
	return &c
}

// ToError returns a pointer to the passed error value.
func ToError(e error) *error {
	return &e
}

// ToFloat32 returns a pointer to the passed float32 value.
func ToFloat32(f float32) *float32 {
	return &f
}

// ToFloat64 returns a pointer to the passed float64 value.
func ToFloat64(f float64) *float64 {
	return &f
}

// ToInt returns a pointer to the passed int value.
func ToInt(i int) *int {
	return &i
}

// ToInt8 returns a pointer to the passed int8 value.
func ToInt8(i int8) *int8 {
	return &i
}

// ToInt16 returns a pointer to the passed int16 value.
func ToInt16(i int16) *int16 {
	return &i
}

// ToInt32 returns a pointer to the passed int32 value.
func ToInt32(i int32) *int32 {
	return &i
}

// ToInt64 returns a pointer to the passed int64 value.
func ToInt64(i int64) *int64 {
	return &i
}

// ToRune returns a pointer to the passed rune value.
func ToRune(r rune) *rune {
	return &r
}

// ToString returns a pointer to the passed string value.
func ToString(s string) *string {
	return &s
}

// ToUint returns a pointer to the passed uint value.
func ToUint(u uint) *uint {
	return &u
}

// ToUint8 returns a pointer to the passed uint8 value.
func ToUint8(u uint8) *uint8 {
	return &u
}

// ToUint16 returns a pointer to the passed uint16 value.
func ToUint16(u uint16) *uint16 {
	return &u
}

// ToUint32 returns a pointer to the passed uint32 value.
func ToUint32(u uint32) *uint32 {
	return &u
}

// ToUint64 returns a pointer to the passed uint64 value.
func ToUint64(u uint64) *uint64 {
	return &u
}

// ToUintptr returns a pointer to the passed uintptr value.
func ToUintptr(u uintptr) *uintptr {
	return &u
}

// ToDuration returns a pointer to the passed time.Duration value.
func ToDuration(d time.Duration) *time.Duration {
	return &d
}

// ToTime returns a pointer to the passed time.Time value.
func ToTime(t time.Time) *time.Time {
	return &t
}
