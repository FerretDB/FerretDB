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

package pgdb

// transactionConflictError is returned when one of the queries in the transaction returned an error because
// of an unexpected conflict. The caller could retry such a transaction.
type transactionConflictError struct {
	err error // underlying error
}

// newTransactionConflictError creates a new transactionConflictError with the given underlying error.
func newTransactionConflictError(err error) error {
	return &transactionConflictError{err: err}
}

// Error implements the error interface.
func (e *transactionConflictError) Error() string {
	return e.err.Error()
}
