package types

type Document struct {
	m    map[string]any
	keys []string
}

type Array struct {
	s []any
}

type Binary struct {
	Subtype BinarySubtype
	B       []byte
}

type BinarySubtype byte

type ObjectID [ObjectIDLen]byte

const ObjectIDLen = 12

type NullType struct{}

type Regex struct {
	Pattern string
	Options string
}

type Timestamp int64
