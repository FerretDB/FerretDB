package pointer

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

// GetBool returns the value of the bool pointer passed in or false if the pointer is nil.
func GetBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// GetByte returns the value of the byte pointer passed in or 0 if the pointer is nil.
func GetByte(b *byte) byte {
	if b == nil {
		return 0
	}
	return *b
}

// GetComplex64 returns the value of the complex64 pointer passed in or 0 if the pointer is nil.
func GetComplex64(c *complex64) complex64 {
	if c == nil {
		return 0
	}
	return *c
}

// GetComplex128 returns the value of the complex128 pointer passed in or 0 if the pointer is nil.
func GetComplex128(c *complex128) complex128 {
	if c == nil {
		return 0
	}
	return *c
}

// GetError returns the value of the error pointer passed in or nil if the pointer is nil.
func GetError(e *error) error {
	if e == nil {
		return nil
	}
	return *e
}

// GetFloat32 returns the value of the float32 pointer passed in or 0 if the pointer is nil.
func GetFloat32(f *float32) float32 {
	if f == nil {
		return 0
	}
	return *f
}

// GetFloat64 returns the value of the float64 pointer passed in or 0 if the pointer is nil.
func GetFloat64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

// GetInt returns the value of the int pointer passed in or 0 if the pointer is nil.
func GetInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// GetInt8 returns the value of the int8 pointer passed in or 0 if the pointer is nil.
func GetInt8(i *int8) int8 {
	if i == nil {
		return 0
	}
	return *i
}

// GetInt16 returns the value of the int16 pointer passed in or 0 if the pointer is nil.
func GetInt16(i *int16) int16 {
	if i == nil {
		return 0
	}
	return *i
}

// GetInt32 returns the value of the int32 pointer passed in or 0 if the pointer is nil.
func GetInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

// GetInt64 returns the value of the int64 pointer passed in or 0 if the pointer is nil.
func GetInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// GetRune returns the value of the rune pointer passed in or 0 if the pointer is nil.
func GetRune(r *rune) rune {
	if r == nil {
		return 0
	}
	return *r
}

// GetString returns the value of the string pointer passed in or empty string if the pointer is nil.
func GetString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// GetUint returns the value of the uint pointer passed in or 0 if the pointer is nil.
func GetUint(u *uint) uint {
	if u == nil {
		return 0
	}
	return *u
}

// GetUint8 returns the value of the uint8 pointer passed in or 0 if the pointer is nil.
func GetUint8(u *uint8) uint8 {
	if u == nil {
		return 0
	}
	return *u
}

// GetUint16 returns the value of the uint16 pointer passed in or 0 if the pointer is nil.
func GetUint16(u *uint16) uint16 {
	if u == nil {
		return 0
	}
	return *u
}

// GetUint32 returns the value of the uint32 pointer passed in or 0 if the pointer is nil.
func GetUint32(u *uint32) uint32 {
	if u == nil {
		return 0
	}
	return *u
}

// GetUint64 returns the value of the uint64 pointer passed in or 0 if the pointer is nil.
func GetUint64(u *uint64) uint64 {
	if u == nil {
		return 0
	}
	return *u
}

// GetUintptr returns the value of the uintptr pointer passed in or 0 if the pointer is nil.
func GetUintptr(u *uintptr) uintptr {
	if u == nil {
		return 0
	}
	return *u
}

// GetDuration returns the value of the duration pointer passed in or 0 if the pointer is nil.
func GetDuration(d *time.Duration) time.Duration {
	if d == nil {
		return 0
	}
	return *d
}

// GetTime returns the value of the time pointer passed in or zero time.Time if the pointer is nil.
func GetTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
