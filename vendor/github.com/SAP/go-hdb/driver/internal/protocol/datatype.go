package protocol

import (
	"database/sql"
	"reflect"
	"time"

	hdbreflect "github.com/SAP/go-hdb/driver/internal/reflect"
)

// DataType is the type definition for data types supported by this package.
type DataType byte

// Data type constants.
const (
	DtUnknown DataType = iota // unknown data type
	DtBoolean
	DtTinyint
	DtSmallint
	DtInteger
	DtBigint
	DtReal
	DtDouble
	DtDecimal
	DtTime
	DtString
	DtBytes
	DtLob
	DtRows
)

// RegisterScanType registers driver owned datatype scantypes (e.g. Decimal, Lob).
func RegisterScanType(dt DataType, scanType, scanNullType reflect.Type) bool {
	scanTypes[dt].scanType = scanType
	scanTypes[dt].scanNullType = scanNullType
	return true
}

var scanTypes = []struct {
	scanType     reflect.Type
	scanNullType reflect.Type
}{
	DtUnknown:  {hdbreflect.TypeFor[any](), hdbreflect.TypeFor[any]()},
	DtBoolean:  {hdbreflect.TypeFor[bool](), hdbreflect.TypeFor[sql.NullBool]()},
	DtTinyint:  {hdbreflect.TypeFor[uint8](), hdbreflect.TypeFor[sql.NullByte]()},
	DtSmallint: {hdbreflect.TypeFor[int16](), hdbreflect.TypeFor[sql.NullInt16]()},
	DtInteger:  {hdbreflect.TypeFor[int32](), hdbreflect.TypeFor[sql.NullInt32]()},
	DtBigint:   {hdbreflect.TypeFor[int64](), hdbreflect.TypeFor[sql.NullInt64]()},
	DtReal:     {hdbreflect.TypeFor[float32](), hdbreflect.TypeFor[sql.NullFloat64]()},
	DtDouble:   {hdbreflect.TypeFor[float64](), hdbreflect.TypeFor[sql.NullFloat64]()},
	DtTime:     {hdbreflect.TypeFor[time.Time](), hdbreflect.TypeFor[sql.NullTime]()},
	DtString:   {hdbreflect.TypeFor[string](), hdbreflect.TypeFor[sql.NullString]()},
	DtBytes:    {nil, nil}, // to be registered by driver
	DtDecimal:  {nil, nil}, // to be registered by driver
	DtLob:      {nil, nil}, // to be registered by driver
	DtRows:     {hdbreflect.TypeFor[sql.Rows](), hdbreflect.TypeFor[sql.Rows]()},
}

// ScanType return the scan type (reflect.Type) of the corresponding data type.
func (dt DataType) ScanType(nullable bool) reflect.Type {
	if nullable {
		return scanTypes[dt].scanNullType
	}
	return scanTypes[dt].scanType
}
