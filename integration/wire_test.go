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

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/FerretDB/wire"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

// TODO https://github.com/FerretDB/FerretDB/issues/344
func FuzzWire(f *testing.F) {
	var limit int
	if testing.Short() {
		limit = 100
	}

	records, err := wire.LoadRecords(testutil.TmpRecordsDir, limit)
	require.NoError(f, err)

	for _, rec := range records {
		b := make([]byte, 0, len(rec.HeaderB)+len(rec.BodyB))
		if cap(b) == 0 {
			continue
		}

		b = append(b, rec.HeaderB...)
		b = append(b, rec.BodyB...)
		f.Add(b)
	}

	f.Logf("%d recorded messages were added to the seed corpus", len(records))

	dbName := testutil.DatabaseName(f)

	f.Fuzz(func(t *testing.T, b []byte) {
		t.Parallel()

		s := setup.SetupWithOpts(t, &setup.SetupOpts{
			DatabaseName: dbName,
			WireConn:     setup.WireConnNoAuth,
		})

		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()

		// we only care about panics (and data races)
		_ = s.WireConn.WriteRaw(ctx, b)
		_, _, _ = s.WireConn.Read(ctx)
	})
}
