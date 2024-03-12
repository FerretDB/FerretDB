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

package wire

import (
	"testing"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

var queryTestCases = []testCase{
	{
		name:    "handshake1",
		headerB: testutil.MustParseDumpFile("testdata", "handshake1_header.hex"),
		bodyB:   testutil.MustParseDumpFile("testdata", "handshake1_body.hex"),
		msgHeader: &MsgHeader{
			MessageLength: 372,
			RequestID:     1,
			ResponseTo:    0,
			OpCode:        OpCodeQuery,
		},
		msgBody: &OpQuery{
			Flags:              0,
			FullCollectionName: "admin.$cmd",
			NumberToSkip:       0,
			NumberToReturn:     -1,
			query: makeRawDocument(
				"ismaster", true,
				"client", must.NotFail(types.NewDocument(
					"driver", must.NotFail(types.NewDocument(
						"name", "nodejs",
						"version", "4.0.0-beta.6",
					)),
					"os", must.NotFail(types.NewDocument(
						"type", "Darwin",
						"name", "darwin",
						"architecture", "x64",
						"version", "20.6.0",
					)),
					"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
					"application", must.NotFail(types.NewDocument(
						"name", "mongosh 1.0.1",
					)),
				)),
				"compression", must.NotFail(types.NewArray("none")),
				"loadBalanced", false,
			),
			returnFieldsSelector: nil,
		},
		m: `
		{
		  "Flags": "[]",
		  "FullCollectionName": "admin.$cmd",
		  "NumberToSkip": 0,
		  "NumberToReturn": -1,
		  "Query": {
		    "ismaster": true,
		    "client": {
		      "driver": {"name": "nodejs", "version": "4.0.0-beta.6"},
		      "os": {
		        "type": "Darwin",
		        "name": "darwin",
		        "architecture": "x64",
		        "version": "20.6.0",
		      },
		      "platform": "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
		      "application": {"name": "mongosh 1.0.1"},
		    },
		    "compression": ["none"],
		    "loadBalanced": false,
		  },
		}`,
	},
	{
		name:    "handshake3",
		headerB: testutil.MustParseDumpFile("testdata", "handshake3_header.hex"),
		bodyB:   testutil.MustParseDumpFile("testdata", "handshake3_body.hex"),
		msgHeader: &MsgHeader{
			MessageLength: 372,
			RequestID:     2,
			ResponseTo:    0,
			OpCode:        OpCodeQuery,
		},
		msgBody: &OpQuery{
			Flags:              0,
			FullCollectionName: "admin.$cmd",
			NumberToSkip:       0,
			NumberToReturn:     -1,
			query: makeRawDocument(
				"ismaster", true,
				"client", must.NotFail(types.NewDocument(
					"driver", must.NotFail(types.NewDocument(
						"name", "nodejs",
						"version", "4.0.0-beta.6",
					)),
					"os", must.NotFail(types.NewDocument(
						"type", "Darwin",
						"name", "darwin",
						"architecture", "x64",
						"version", "20.6.0",
					)),
					"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
					"application", must.NotFail(types.NewDocument(
						"name", "mongosh 1.0.1",
					)),
				)),
				"compression", must.NotFail(types.NewArray("none")),
				"loadBalanced", false,
			),
			returnFieldsSelector: nil,
		},
		m: `
		{
		  "Flags": "[]",
		  "FullCollectionName": "admin.$cmd",
		  "NumberToSkip": 0,
		  "NumberToReturn": -1,
		  "Query": {
		    "ismaster": true,
		    "client": {
		      "driver": {"name": "nodejs", "version": "4.0.0-beta.6"},
		      "os": {
		        "type": "Darwin",
		        "name": "darwin",
		        "architecture": "x64",
		        "version": "20.6.0",
		      },
		      "platform": "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
		      "application": {"name": "mongosh 1.0.1"},
		    },
		    "compression": ["none"],
		    "loadBalanced": false,
		  },
		}`,
	},
}

func TestQuery(t *testing.T) {
	t.Parallel()
	testMessages(t, queryTestCases)
}

func FuzzQuery(f *testing.F) {
	fuzzMessages(f, queryTestCases)
}
