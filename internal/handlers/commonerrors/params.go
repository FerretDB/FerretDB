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

package commonerrors

import "fmt"

var (
	// ErrUnexpectedType is returned when the type of a value is unexpected.
	ErrUnexpectedType = fmt.Errorf("unexpected type")
	// ErrNotWholeNumber is returned when a number is not a whole number.
	ErrNotWholeNumber = fmt.Errorf("not a whole number")
	// ErrNegativeNumber is returned when a number is negative.
	ErrNegativeNumber = fmt.Errorf("negative number")
	// ErrNotBinaryMask is returned when a number is not a binary mask.
	ErrNotBinaryMask = fmt.Errorf("not a binary mask")
	// ErrUnexpectedLeftOpType is returned when the type of the left operand is unexpected.
	ErrUnexpectedLeftOpType = fmt.Errorf("unexpected left operand type")
	// ErrUnexpectedRightOpType is returned when the type of the right operand is unexpected.
	ErrUnexpectedRightOpType = fmt.Errorf("unexpected right operand type")
	// ErrLongExceededPositive is returned when a positive long value exceeds the maximum value.
	ErrLongExceededPositive = fmt.Errorf("long exceeded - positive value")
	// ErrLongExceededNegative is returned when a negative long value exceeds the minimum value.
	ErrLongExceededNegative = fmt.Errorf("long exceeded - negative value")
	// ErrIntExceeded is returned when an int value exceeds the maximum value.
	ErrIntExceeded = fmt.Errorf("int exceeded")
	// ErrInfinity is returned when a value is infinity.
	ErrInfinity = fmt.Errorf("infinity")
)
