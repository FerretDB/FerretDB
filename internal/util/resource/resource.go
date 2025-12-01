// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package resource provides utilities for tracking resource lifetimes.
package resource

import (
	"fmt"
	"reflect"
	"runtime"
	runtimedebug "runtime/debug"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/FerretDB/FerretDB/v2/internal/util/devbuild"
)

// Token is a field of a tracked object, holding the cleanup handle and a cleanup function with message.
type Token struct {
	h       atomic.Pointer[runtime.Cleanup]
	cleanup func(*Token)
	msg     string
}

// cleanup panics with the token's message.
func cleanup(t *Token) {
	panic(t.msg)
}

// NewToken returns a new Token.
func NewToken() *Token {
	return &Token{
		cleanup: cleanup,
	}
}

// profilesM protects access to profiles.
var profilesM sync.Mutex

// profileName return pprof profile name for the given object.
func profileName(obj any) string {
	return "FerretDB/" + reflect.TypeOf(obj).Elem().String()
}

// Track tracks the lifetime of an object until Untrack is called on it.
//
// Obj should be a pointer to a struct with a field "token" of type *Token.
func Track[T any](obj *T, token *Token) {
	checkArgs(obj, token)

	name := profileName(obj)

	// fast path

	p := pprof.Lookup(name)

	if p == nil {
		// slow path

		profilesM.Lock()

		// a concurrent call might have created a profile already; check again
		if p = pprof.Lookup(name); p == nil {
			p = pprof.NewProfile(name)
		}

		profilesM.Unlock()
	}

	// use token instead of obj itself,
	// because otherwise profile will hold a reference to obj and cleanup will never run
	p.Add(token, 1)

	token.msg = fmt.Sprintf("%T has not been finalized", obj)
	if devbuild.Enabled {
		token.msg += "\nObject created by " + string(runtimedebug.Stack())
	}

	h := runtime.AddCleanup(obj, token.cleanup, token)
	token.h.Store(&h)
}

// Untrack stops tracking the lifetime of an object.
//
// It is safe to call this function multiple times concurrently.
func Untrack[T any](obj *T, token *Token) {
	checkArgs(obj, token)

	if h := token.h.Swap(nil); h != nil {
		h.Stop()
	}

	p := pprof.Lookup(profileName(obj))
	if p == nil {
		panic("object is not tracked")
	}

	p.Remove(token)
}

// checkArgs checks Track and Untrack arguments.
//
// Other creative misuses of Track should result in panics too, if less clear.
func checkArgs(obj any, token *Token) {
	if obj == nil {
		panic("obj must not be nil")
	}

	if token == nil {
		panic("token must not be nil")
	}

	pv := reflect.ValueOf(obj)
	if pv.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("obj must be a pointer to struct, got %T", obj))
	}

	v := pv.Elem()
	if v.Kind() != reflect.Struct {
		panic(fmt.Sprintf("obj must be a pointer to struct, got %T", obj))
	}

	f := v.FieldByName("token")
	if f.Kind() != reflect.Ptr || f.UnsafePointer() != unsafe.Pointer(token) {
		panic("token must be a pointer field of a struct")
	}
}
