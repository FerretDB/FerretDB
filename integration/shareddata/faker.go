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
	"math"
	"math/rand"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// words is a list of words used for generating random data.
var words = [...]string{
	"access",
	"additional",
	"against",
	"aggregate",
	"all",
	"allow",
	"allowed",
	"allows",
	"along",
	"an",
	"and",
	"any",
	"anyone",
	"apply",
	"are",
	"as",
	"at",
	"attached",
	"author",
	"away",
	"be",
	"being",
	"build",
	"built",
	"business",
	"by",
	"carry",
	"charge",
	"code",
	"compiled",
	"comply",
	"component",
	"conjunction",
	"containing",
	"cost",
	"criteria",
	"deliberately",
	"depend",
	"derived",
	"different",
	"discriminate",
	"discrimination",
	"distributed",
	"distribution",
	"doesn",
	"downloading",
	"endeavor",
	"example",
	"execution",
	"explicitly",
	"extracted",
	"fee",
	"field",
	"fields",
	"files”",
	"following",
	"for",
	"form",
	"forms",
	"free",
	"from",
	"genetic",
	"giving",
	"granted",
	"group",
	"groups",
	"have",
	"if",
	"in",
	"include",
	"individual",
	"insist",
	"integrity",
	"interface",
	"intermediate",
	"internet",
	"introduction",
	"is",
	"it",
	"just",
	"license",
	"licensed",
	"making",
	"may",
	"mean",
	"means",
	"medium",
	"modifications",
	"modified",
	"modify",
	"modifying",
	"more",
	"must",
	"name",
	"need",
	"no",
	"not",
	"number",
	"obfuscated",
	"obtaining",
	"of",
	"on",
	"only",
	"open",
	"or",
	"original",
	"other",
	"output",
	"part",
	"particular",
	"parties",
	"party",
	"patch",
	"permit",
	"person",
	"persons",
	"place",
	"predicated",
	"preferably",
	"preferred",
	"preprocessor",
	"product",
	"program",
	"programmer",
	"programs",
	"provision",
	"purpose",
	"reasonable",
	"redistributed",
	"redistribution",
	"reproduction",
	"require",
	"research",
	"restrict",
	"restrictions",
	"rights",
	"royalty",
	"sale",
	"same",
	"selling",
	"several",
	"shall",
	"should",
	"software",
	"some",
	"source",
	"source-code",
	"sources",
	"specific",
	"style",
	"such",
	"technology",
	"technology-neutral",
	"terms",
	"than",
	"that",
	"the",
	"them",
	"there",
	"those",
	"time",
	"to",
	"translator",
	"under",
	"use",
	"used",
	"version",
	"via",
	"well",
	"well-publicized",
	"where",
	"which",
	"whom",
	"with",
	"within",
	"without",
	"works",
	"would",
}

// wordsB is words as []byte.
var wordsB [len(words)][]byte

// init initializes wordsB.
func init() {
	for i, w := range words {
		wordsB[i] = []byte(w)
	}
}

// faker generates stable fake data for tests and benchmarks.
//
// It is not thread-safe.
type faker struct {
	r *rand.Rand
}

// newFaker creates a new faker.
func newFaker() *faker {
	src := rand.NewSource(1)

	return &faker{
		r: rand.New(src),
	}
}

// FieldName generates a random document field name.
func (f *faker) FieldName() string {
	return "field_" + strconv.Itoa(f.r.Int())
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
			for {
				f := math.Float64frombits(f.r.Uint64())
				if !math.IsNaN(f) { // to simplify comparisons
					return f
				}
			}
		case 0x02: // String
			return words[f.r.Intn(len(words))]
		case 0x05: // Binary
			return primitive.Binary{Subtype: 0x00, Data: wordsB[f.r.Intn(len(wordsB))]}
		case 0x07: // ObjectID
			return f.ObjectID()
		case 0x08: // Bool
			return f.r.Intn(2) == 1
		case 0x09: // DateTime
			return time.UnixMilli(int64(int32(f.r.Uint32())) * 1000) // 1901-2038 to make JSON encoder happy
		case 0x0a: // Null
			return nil
		case 0x0b: // Regex
			// not yet
		case 0x10: // Int32
			return int32(f.r.Uint32())
		case 0x11: // Timestamp
			// not yet
		case 0x12: // Int64
			return int64(f.r.Uint64())
		case 0x13: // Decimal
			// TODO https://github.com/FerretDB/FerretDB/issues/66
		default:
		}
	}
}
