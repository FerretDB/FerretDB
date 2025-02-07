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

package admin

import (
	"math"
	"strings"
	"sync"
	"testing"

	"github.com/FerretDB/wire/wirebson"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/integration"
	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestKillOp(t *testing.T) {
	ctx, collection := setup.Setup(t)
	adminDB := collection.Database().Client().Database("admin")

	searchF := func(op *wirebson.Document) bool {
		return op.Get("op") == "insert" && op.Get("ns") == collection.Name()+"."+collection.Database().Name()
	}

	var docs []any
	for range 3 {
		docs = append(docs, bson.D{
			{"strings", strings.Repeat("something for testing", 500)},
		})
	}

	var wg sync.WaitGroup

	nOps := 10
	nCurrentOp := 250
	attempts := 1

	var opErr error

	for range attempts {
		readyCh := make(chan struct{}, nOps+nCurrentOp)
		startCh := make(chan struct{})
		opErrCh := make(chan error, nOps)

		for i := 0; i < nOps; i++ {
			wg.Add(1)

			go func() {
				wg.Done()

				readyCh <- struct{}{}

				<-startCh

				if _, err := collection.InsertMany(ctx, docs); err != nil {
					opErrCh <- err
				}
			}()
		}

		for i := 0; i < nCurrentOp; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				readyCh <- struct{}{}

				<-startCh

				var actual bson.D
				err := adminDB.RunCommand(ctx, bson.D{{"currentOp", 1}}).Decode(&actual)
				require.NoError(t, err)

				doc, err := integration.Convert(t, actual).(wirebson.AnyDocument).Decode()
				require.NoError(t, err)

				v := doc.Get("inprog")
				require.NotNil(t, v)

				inprog, err := v.(wirebson.AnyArray).Decode()
				require.NoError(t, err)

				for v = range inprog.Values() {
					var op *wirebson.Document

					op, err = v.(wirebson.AnyDocument).Decode()
					require.NoError(t, err)

					if searchF(op) {
						opid := op.Get("opid")

						var res bson.D
						err = adminDB.RunCommand(ctx, bson.D{
							{"killOp", int32(1)},
							{"op", opid},
						}).Decode(&res)
						require.NoError(t, err)

						expected := bson.D{
							{"info", "attempting to kill op"},
							{"ok", float64(1)},
						}
						integration.AssertEqualDocuments(t, expected, res)
					}
				}
			}()
		}

		for i := 0; i < len(readyCh); i++ {
			if _, ok := <-readyCh; !ok {
				t.Error("closed ready channel")
			}
		}

		close(startCh)

		wg.Wait()

		var ok bool
		if opErr, ok = <-opErrCh; ok {
			break
		}
	}

	// MongoDB error "bulk write exception: write concern error: (Interrupted) operation was interrupted"
	// FerretDB error "(InternalError) timeout: context already done: context canceled"
	require.Error(t, opErr)

	count, err := collection.CountDocuments(ctx, bson.D{})
	require.NoError(t, err)
	require.Less(t, count, int64(len(docs))*int64(nOps))
}

func TestKillOpNonExistentOp(t *testing.T) {
	ctx, collection := setup.Setup(t)
	adminDB := collection.Database().Client().Database("admin")

	var res bson.D
	err := adminDB.RunCommand(ctx, bson.D{
		{"killOp", int32(1)},
		{"op", int64(112233)},
	}).Decode(&res)
	require.NoError(t, err)

	expected := bson.D{
		{"info", "attempting to kill op"},
		{"ok", float64(1)},
	}
	integration.AssertEqualDocuments(t, expected, res)
}

func TestKillOpErrors(t *testing.T) {
	ctx, collection := setup.Setup(t)
	adminDB := collection.Database().Client().Database("admin")

	for name, tc := range map[string]struct { //nolint:vet // test only
		db      *mongo.Database // defaults to adminDB
		command bson.D

		err        *mongo.CommandError
		altMessage string
	}{
		"NonAdminDB": {
			db: collection.Database(),
			command: bson.D{
				{"killOp", int32(1)},
				{"op", int64(112233)},
			},
			err: &mongo.CommandError{
				Code:    13,
				Name:    "Unauthorized",
				Message: "killOp may only be run against the admin database.",
			},
		},
		"MissingOpID": {
			command: bson.D{
				{"killOp", int32(1)},
			},
			err: &mongo.CommandError{
				Code:    4,
				Name:    "NoSuchKey",
				Message: `Missing expected field "op"`,
			},
		},
		"StringOpID": {
			command: bson.D{
				{"killOp", int32(1)},
				{"op", "invalid"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: `Expected field "op" to have numeric type, but found string`,
			},
		},
		"TooLargeLongOpID": {
			command: bson.D{
				{"killOp", int32(1)},
				{"op", math.MaxInt64},
			},
			err: &mongo.CommandError{
				Code:    26823,
				Name:    "Location26823",
				Message: `invalid op : 9223372036854775807. Op ID cannot be represented with 32 bits`,
			},
		},
		"TooLargeDoubleOpID": {
			command: bson.D{
				{"killOp", int32(1)},
				{"op", math.MaxFloat64},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Expected field "op" to have a value exactly representable as a 64-bit integer, but found op: 1.797693134862316e+308`,
			},
			altMessage: `Expected field "op" to have a value exactly representable as a 64-bit integer, but found op: 1.7976931348623157e+308`,
		},
		"DoubleOpID": {
			command: bson.D{
				{"killOp", int32(1)},
				{"op", 1.1},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Expected field "op" to have a value exactly representable as a 64-bit integer, but found op: 1.1`,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			db := tc.db
			if db == nil {
				db = adminDB
			}

			err := db.RunCommand(ctx, tc.command).Err()
			integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}
