package mongoerrors

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestPostgreSQLConcurrencyErrors(t *testing.T) {
	for _, tc := range []struct {
		state string
		code  int32
		name  string
	}{
		{"40P01", 112, "WriteConflict"},
		{"40001", 112, "WriteConflict"},
		{"55P03", 24, "LockTimeout"},
		{"57014", 50, "MaxTimeMSExpired"},
	} {
		t.Run(tc.state, func(t *testing.T) {
			err := Make(testutil.Ctx(t), &pgconn.PgError{Code: tc.state, Message: "synthetic failure"}, "probe", testutil.Logger(t))
			if err.Code != tc.code || err.Name != tc.name {
				t.Fatalf("unexpected mapped error: %v", err)
			}
		})
	}
}

func TestContextDeadlineMapsToMongoTimeout(t *testing.T) {
	err := Make(testutil.Ctx(t), context.DeadlineExceeded, "probe", testutil.Logger(t))
	if err.Code != 50 || err.Name != "MaxTimeMSExpired" {
		t.Fatalf("unexpected mapped deadline: %v", err)
	}
}
