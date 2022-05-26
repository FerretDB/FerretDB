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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either expequals or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package integration provides FerretDB integration tests.
package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestAssertEqualWriteError(t *testing.T) {
	t.Parallel()

	type testCase struct {
		err1  *mongo.WriteError
		alt   string
		err2  mongo.WriteError
		equal bool
	}

	for name, tc := range map[string]testCase{
		"Equal": {
			err1: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type string instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: \"string\"}",
			},
			err2: mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type string instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: \"string\"}",
			},
			equal: true,
		},
		"Differ": {
			err1: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: 42.12345}",
			},
			err2: mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: 42.12}",
			},
			equal: false,
		},
		"DifferAlt": {
			err1: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: 42.12345}",
			},
			alt: "Modifiers operate on fields but we found type double instead. " +
				"For example: {$mod: {<field>: ...}} not {$set: 42.12}",
			err2: mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: 42.12}",
			},
			equal: true,
		},
		"DifferCodes": {
			err1: &mongo.WriteError{
				Code: 29,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: 42.12345}",
			},
			alt: "Modifiers operate on fields but we found type double instead. " +
				"For example: {$mod: {<field>: ...}} not {$set: 42.12}",
			err2: mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: 42.12}",
			},
			equal: false,
		},
	} {
		equal := AssertEqualWriteError(t, tc.err1, tc.alt, tc.err2)
		assert.Equal(t, tc.equal, equal, name)
	}
}
