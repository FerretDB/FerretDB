//go:build go1.22

// Package reflect provides a compatibility layer until g1.21 is out of maintenance.
package reflect

import "reflect"

// TypeFor returns the [Type] that represents the type argument T.
func TypeFor[T any]() reflect.Type {
	return reflect.TypeFor[T]()
}
