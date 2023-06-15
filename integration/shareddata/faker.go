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

package shareddata

import (
	"fmt"
	"math/rand"
	"time"

	fakerlib "github.com/jaswdr/faker"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// faker generates stable fake data for tests.
//
// It is not thread-safe.
type faker struct {
	r *rand.Rand
	f fakerlib.Faker
}

// newFaker creates a new faker.
func newFaker() *faker {
	src := rand.NewSource(1)
	return &faker{
		r: rand.New(src),
		f: fakerlib.NewWithSeed(src),
	}
}

// FieldName generates a random document field name.
func (f *faker) FieldName() string {
	// TODO https://github.com/jaswdr/faker/issues/142

	// somewhat surprisingly, generating "better" field names generates duplicates too often
	return fmt.Sprintf("field_%d", f.r.Int())
}

// ObjectID generates a random non-zero ObjectID.
func (f *faker) ObjectID() primitive.ObjectID {
	var id primitive.ObjectID
	for id.IsZero() {
		must.NotFail(f.r.Read(id[:]))
	}
	return id
}

// ScalarValue generates a random scalar value.
func (f *faker) ScalarValue() any {
	for {
		switch f.r.Intn(0x13) + 1 {
		case 0x01: // Double
			return f.r.Float64()
		case 0x02: // String
			return f.f.Lorem().Words(10)
		case 0x05: // Binary
			return primitive.Binary{Subtype: 0x00, Data: f.f.Lorem().Bytes(1000)}
		case 0x07: // ObjectID
			return f.ObjectID()
		case 0x08: // Bool
			return f.f.Bool()
		case 0x09: // DateTime
			return f.f.Time().Time(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))
		case 0x0a: // Null
			return nil
		case 0x0b: // Regex
			// TODO
		case 0x10: // Int32
			return f.f.Int32()
		case 0x11: // Timestamp
			// TODO
		case 0x12: // Int64
			return f.f.Int64()
		case 0x13: // Decimal
			// TODO https://github.com/FerretDB/FerretDB/issues/66
		default:
		}
	}
}
