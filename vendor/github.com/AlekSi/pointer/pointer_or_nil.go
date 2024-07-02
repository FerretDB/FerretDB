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

// ToBoolOrNil returns a pointer to the passed bool value, or nil, if passed value is a zero value.
func ToBoolOrNil(b bool) *bool {
	if b == false {
		return nil
	}
	return &b
}

// ToByteOrNil returns a pointer to the passed byte value, or nil, if passed value is a zero value.
func ToByteOrNil(b byte) *byte {
	if b == 0 {
		return nil
	}
	return &b
}

// ToComplex64OrNil returns a pointer to the passed complex64 value, or nil, if passed value is a zero value.
func ToComplex64OrNil(c complex64) *complex64 {
	if c == 0 {
		return nil
	}
	return &c
}

// ToComplex128OrNil returns a pointer to the passed complex128 value, or nil, if passed value is a zero value.
func ToComplex128OrNil(c complex128) *complex128 {
	if c == 0 {
		return nil
	}
	return &c
}

// ToErrorOrNil returns a pointer to the passed error value, or nil, if passed value is a zero value.
func ToErrorOrNil(e error) *error {
	if e == nil {
		return nil
	}
	return &e
}

// ToFloat32OrNil returns a pointer to the passed float32 value, or nil, if passed value is a zero value.
func ToFloat32OrNil(f float32) *float32 {
	if f == 0 {
		return nil
	}
	return &f
}

// ToFloat64OrNil returns a pointer to the passed float64 value, or nil, if passed value is a zero value.
func ToFloat64OrNil(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

// ToIntOrNil returns a pointer to the passed int value, or nil, if passed value is a zero value.
func ToIntOrNil(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

// ToInt8OrNil returns a pointer to the passed int8 value, or nil, if passed value is a zero value.
func ToInt8OrNil(i int8) *int8 {
	if i == 0 {
		return nil
	}
	return &i
}

// ToInt16OrNil returns a pointer to the passed int16 value, or nil, if passed value is a zero value.
func ToInt16OrNil(i int16) *int16 {
	if i == 0 {
		return nil
	}
	return &i
}

// ToInt32OrNil returns a pointer to the passed int32 value, or nil, if passed value is a zero value.
func ToInt32OrNil(i int32) *int32 {
	if i == 0 {
		return nil
	}
	return &i
}

// ToInt64OrNil returns a pointer to the passed int64 value, or nil, if passed value is a zero value.
func ToInt64OrNil(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

// ToRuneOrNil returns a pointer to the passed rune value, or nil, if passed value is a zero value.
func ToRuneOrNil(r rune) *rune {
	if r == 0 {
		return nil
	}
	return &r
}

// ToStringOrNil returns a pointer to the passed string value, or nil, if passed value is a zero value.
func ToStringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ToUintOrNil returns a pointer to the passed uint value, or nil, if passed value is a zero value.
func ToUintOrNil(u uint) *uint {
	if u == 0 {
		return nil
	}
	return &u
}

// ToUint8OrNil returns a pointer to the passed uint8 value, or nil, if passed value is a zero value.
func ToUint8OrNil(u uint8) *uint8 {
	if u == 0 {
		return nil
	}
	return &u
}

// ToUint16OrNil returns a pointer to the passed uint16 value, or nil, if passed value is a zero value.
func ToUint16OrNil(u uint16) *uint16 {
	if u == 0 {
		return nil
	}
	return &u
}

// ToUint32OrNil returns a pointer to the passed uint32 value, or nil, if passed value is a zero value.
func ToUint32OrNil(u uint32) *uint32 {
	if u == 0 {
		return nil
	}
	return &u
}

// ToUint64OrNil returns a pointer to the passed uint64 value, or nil, if passed value is a zero value.
func ToUint64OrNil(u uint64) *uint64 {
	if u == 0 {
		return nil
	}
	return &u
}

// ToUintptrOrNil returns a pointer to the passed uintptr value, or nil, if passed value is a zero value.
func ToUintptrOrNil(u uintptr) *uintptr {
	if u == 0 {
		return nil
	}
	return &u
}

// ToDurationOrNil returns a pointer to the passed time.Duration value, or nil, if passed value is a zero value.
func ToDurationOrNil(d time.Duration) *time.Duration {
	if d == 0 {
		return nil
	}
	return &d
}

// ToTimeOrNil returns a pointer to the passed time.Time value, or nil, if passed value is a zero value (t.IsZero() returns true).
func ToTimeOrNil(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
