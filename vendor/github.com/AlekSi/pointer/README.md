# pointer

[![Go Reference](https://pkg.go.dev/badge/github.com/AlekSi/pointer.svg)](https://pkg.go.dev/github.com/AlekSi/pointer)

Go package `pointer` provides helpers to convert between pointers and values of built-in
(and, with Go 1.18+ generics, of any) types.

```
go get github.com/AlekSi/pointer
```

API is stable. [Documentation](https://pkg.go.dev/github.com/AlekSi/pointer).

```go
package motivationalexample

import (
	"encoding/json"

	"github.com/AlekSi/pointer"
)

const (
	defaultName = "some name"
)

// Stuff contains optional fields.
type Stuff struct {
	Name    *string
	Comment *string
	Value   *int64
	Time    *time.Time
}

// SomeStuff makes some JSON-encoded stuff.
func SomeStuff() (data []byte, err error) {
	return json.Marshal(&Stuff{
		Name:    pointer.ToString(defaultName),                                   // can't say &defaultName
		Comment: pointer.ToString("not yet"),                                     // can't say &"not yet"
		Value:   pointer.ToInt64(42),                                             // can't say &42 or &int64(42)
		Time:    pointer.ToTime(time.Date(2014, 6, 25, 12, 24, 40, 0, time.UTC)), // can't say &time.Date(…)
	})
}
```
