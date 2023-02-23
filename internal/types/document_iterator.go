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

package types

import (
	"runtime"
	"runtime/pprof"
	"sync/atomic"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// documentIteratorProfiles keeps track on all document iterators.
var documentIteratorProfiles = pprof.NewProfile("github.com/FerretDB/FerretDB/internal/types.documentIterator")

// documentIterator represents an iterator over the document fields.
type documentIterator struct {
	n     atomic.Uint32
	doc   *Document
	stack []byte
}

// newDocumentIterator creates a new document iterator.
func newDocumentIterator(document *Document) iterator.Interface[string, any] {
	iter := &documentIterator{
		doc:   document,
		stack: debugbuild.Stack(),
	}

	documentIteratorProfiles.Add(iter, 1)

	runtime.SetFinalizer(iter, func(iter *documentIterator) {
		msg := "documentIterator.Close() has not been called"
		if iter.stack != nil {
			msg += "\ndocumentIterator created by " + string(iter.stack)
		}

		panic(msg)
	})

	return iter
}

// Next implements iterator.Interface.
func (iter *documentIterator) Next() (string, any, error) {
	n := int(iter.n.Add(1)) - 1

	if n >= iter.doc.Len() {
		return "", nil, iterator.ErrIteratorDone
	}

	return iter.doc.fields[n].key, iter.doc.fields[n].value, nil
}

// Close implements iterator.Interface.
func (iter *documentIterator) Close() {
	iter.n.Store(uint32(iter.doc.Len()))

	documentIteratorProfiles.Remove(iter)

	runtime.SetFinalizer(iter, nil)
}

// check interfaces
var (
	_ iterator.Interface[string, any] = (*documentIterator)(nil)
)
