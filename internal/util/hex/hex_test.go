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

package hex

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const wiresharkDump = `
0000   74 01 00 00 01 00 00 00 00 00 00 00 d4 07 00 00   t...............
0010   00 00 00 00 61 64 6d 69 6e 2e 24 63 6d 64 00 00   ....admin.$cmd..
0020   00 00 00 ff ff ff ff 4d 01 00 00 08 69 73 6d 61   .......M....isma
0030   73 74 65 72 00 01 03 63 6c 69 65 6e 74 00 08 01   ster...client...
0040   00 00 03 64 72 69 76 65 72 00 30 00 00 00 02 6e   ...driver.0....n
0050   61 6d 65 00 07 00 00 00 6e 6f 64 65 6a 73 00 02   ame.....nodejs..
0060   76 65 72 73 69 6f 6e 00 0d 00 00 00 34 2e 30 2e   version.....4.0.
0070   30 2d 62 65 74 61 2e 36 00 00 03 6f 73 00 51 00   0-beta.6...os.Q.
0080   00 00 02 74 79 70 65 00 07 00 00 00 44 61 72 77   ...type.....Darw
0090   69 6e 00 02 6e 61 6d 65 00 07 00 00 00 64 61 72   in..name.....dar
00a0   77 69 6e 00 02 61 72 63 68 69 74 65 63 74 75 72   win..architectur
00b0   65 00 04 00 00 00 78 36 34 00 02 76 65 72 73 69   e.....x64..versi
00c0   6f 6e 00 07 00 00 00 32 30 2e 36 2e 30 00 00 02   on.....20.6.0...
00d0   70 6c 61 74 66 6f 72 6d 00 3e 00 00 00 4e 6f 64   platform.>...Nod
00e0   65 2e 6a 73 20 76 31 34 2e 31 37 2e 33 2c 20 4c   e.js v14.17.3, L
00f0   45 20 28 75 6e 69 66 69 65 64 29 7c 4e 6f 64 65   E (unified)|Node
0100   2e 6a 73 20 76 31 34 2e 31 37 2e 33 2c 20 4c 45   .js v14.17.3, LE
0110   20 28 75 6e 69 66 69 65 64 29 00 03 61 70 70 6c    (unified)..appl
0120   69 63 61 74 69 6f 6e 00 1d 00 00 00 02 6e 61 6d   ication......nam
0130   65 00 0e 00 00 00 6d 6f 6e 67 6f 73 68 20 31 2e   e.....mongosh 1.
0140   30 2e 31 00 00 00 04 63 6f 6d 70 72 65 73 73 69   0.1....compressi
0150   6f 6e 00 11 00 00 00 02 30 00 05 00 00 00 6e 6f   on......0.....no
0160   6e 65 00 00 08 6c 6f 61 64 42 61 6c 61 6e 63 65   ne...loadBalance
0170   64 00 00 00                                       d...
`

const wiresharkExpected = "" +
	"\x74\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\xd4\x07\x00\x00" +
	"\x00\x00\x00\x00\x61\x64\x6d\x69\x6e\x2e\x24\x63\x6d\x64\x00\x00" +
	"\x00\x00\x00\xff\xff\xff\xff\x4d\x01\x00\x00\x08\x69\x73\x6d\x61" +
	"\x73\x74\x65\x72\x00\x01\x03\x63\x6c\x69\x65\x6e\x74\x00\x08\x01" +
	"\x00\x00\x03\x64\x72\x69\x76\x65\x72\x00\x30\x00\x00\x00\x02\x6e" +
	"\x61\x6d\x65\x00\x07\x00\x00\x00\x6e\x6f\x64\x65\x6a\x73\x00\x02" +
	"\x76\x65\x72\x73\x69\x6f\x6e\x00\x0d\x00\x00\x00\x34\x2e\x30\x2e" +
	"\x30\x2d\x62\x65\x74\x61\x2e\x36\x00\x00\x03\x6f\x73\x00\x51\x00" +
	"\x00\x00\x02\x74\x79\x70\x65\x00\x07\x00\x00\x00\x44\x61\x72\x77" +
	"\x69\x6e\x00\x02\x6e\x61\x6d\x65\x00\x07\x00\x00\x00\x64\x61\x72" +
	"\x77\x69\x6e\x00\x02\x61\x72\x63\x68\x69\x74\x65\x63\x74\x75\x72" +
	"\x65\x00\x04\x00\x00\x00\x78\x36\x34\x00\x02\x76\x65\x72\x73\x69" +
	"\x6f\x6e\x00\x07\x00\x00\x00\x32\x30\x2e\x36\x2e\x30\x00\x00\x02" +
	"\x70\x6c\x61\x74\x66\x6f\x72\x6d\x00\x3e\x00\x00\x00\x4e\x6f\x64" +
	"\x65\x2e\x6a\x73\x20\x76\x31\x34\x2e\x31\x37\x2e\x33\x2c\x20\x4c" +
	"\x45\x20\x28\x75\x6e\x69\x66\x69\x65\x64\x29\x7c\x4e\x6f\x64\x65" +
	"\x2e\x6a\x73\x20\x76\x31\x34\x2e\x31\x37\x2e\x33\x2c\x20\x4c\x45" +
	"\x20\x28\x75\x6e\x69\x66\x69\x65\x64\x29\x00\x03\x61\x70\x70\x6c" +
	"\x69\x63\x61\x74\x69\x6f\x6e\x00\x1d\x00\x00\x00\x02\x6e\x61\x6d" +
	"\x65\x00\x0e\x00\x00\x00\x6d\x6f\x6e\x67\x6f\x73\x68\x20\x31\x2e" +
	"\x30\x2e\x31\x00\x00\x00\x04\x63\x6f\x6d\x70\x72\x65\x73\x73\x69" +
	"\x6f\x6e\x00\x11\x00\x00\x00\x02\x30\x00\x05\x00\x00\x00\x6e\x6f" +
	"\x6e\x65\x00\x00\x08\x6c\x6f\x61\x64\x42\x61\x6c\x61\x6e\x63\x65" +
	"\x64\x00\x00\x00"

const goDump = `
00000000  03 64 72 69 76 65 72 00  30 00 00 00 02 6e 61 6d  |.driver.0....nam|
00000010  65 00 07 00 00 00 6e 6f  64 65 6a 73 00 02 76 65  |e.....nodejs..ve|
00000020  72 73 69 6f 6e 00 0d 00  00 00 34 2e 30 2e 30 2d  |rsion.....4.0.0-|
00000030  62 65 74 61 2e 36 00 00  03 6f 73 00 51 00 00 00  |beta.6...os.Q...|
00000040  02 74 79 70 65 00 07 00  00 00 44 61 72 77 69 6e  |.type.....Darwin|
00000050  00 02 6e 61 6d 65 00 07  00 00 00 64 61 72 77 69  |..name.....darwi|
00000060  6e 00 02 61 72 63 68 69  74 65 63 74 75 72 65 00  |n..architecture.|
00000070  04 00 00 00 78 36 34 00  02 76 65 72 73 69 6f 6e  |....x64..version|
00000080  00 07 00 00 00 32 30 2e  36 2e 30 00 00 02 70 6c  |.....20.6.0...pl|
00000090  61 74 66 6f 72 6d 00 3e  00 00 00 4e 6f 64 65 2e  |atform.>...Node.|
000000a0  6a 73 20 76 31 34 2e 31  37 2e 33 2c 20 4c 45 20  |js v14.17.3, LE |
000000b0  28 75 6e 69 66 69 65 64  29 7c 4e 6f 64 65 2e 6a  |(unified)|Node.j|
000000c0  73 20 76 31 34 2e 31 37  2e 33 2c 20 4c 45 20 28  |s v14.17.3, LE (|
000000d0  75 6e 69 66 69 65 64 29  00 03 61 70 70 6c 69 63  |unified)..applic|
000000e0  61 74 69 6f 6e 00 1d 00  00 00 02 6e 61 6d 65 00  |ation......name.|
000000f0  0e 00 00 00 6d 6f 6e 67  6f 73 68 20 31 2e 30 2e  |....mongosh 1.0.|
00000100  31 00 00 00                                       |1...|
`

const goExpected = `` +
	`036472697665720030000000026e616d6500070000006e6f64656a730002` +
	`76657273696f6e000d000000342e302e302d626574612e360000036f7300` +
	`510000000274797065000700000044617277696e00026e616d6500070000` +
	`0064617277696e0002617263686974656374757265000400000078363400` +
	`0276657273696f6e000700000032302e362e30000002706c6174666f726d` +
	`003e0000004e6f64652e6a73207631342e31372e332c204c452028756e69` +
	`66696564297c4e6f64652e6a73207631342e31372e332c204c452028756e` +
	`69666965642900036170706c69636174696f6e001d000000026e616d6500` +
	`0e0000006d6f6e676f736820312e302e31000000`

func TestParseDump(t *testing.T) {
	t.Parallel()

	actual, err := ParseDump(wiresharkDump)
	require.NoError(t, err)
	assert.Equal(t, []byte(wiresharkExpected), actual)

	actual, err = ParseDump(goDump)
	require.NoError(t, err)
	goExpectedB, err := hex.DecodeString(goExpected)
	require.NoError(t, err)
	assert.Equal(t, goExpectedB, actual)
}
