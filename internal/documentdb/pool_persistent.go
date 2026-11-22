package documentdb

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

// persistCursor accounts for connections detached from pgxpool. Admission is
// nonblocking: waiting while holding a pool connection or namespace read barrier
// could prevent an existing cursor from closing and create a gateway deadlock.
func (p *Pool) persistCursor(ctx context.Context, conn *Conn) (*pgx.Conn, func(), error) {
	if p.persistentSlots == nil {
		return conn.hijack(), nil, nil
	}
	select {
	case p.persistentSlots <- struct{}{}:
		var once sync.Once
		release := func() { once.Do(func() { <-p.persistentSlots }) }
		return conn.hijack(), release, nil
	default:
		// The native first-page call may have opened a server cursor. Closing
		// this still-pooled connection releases it before reporting the quota.
		closing, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		_ = conn.Conn().Close(closing)
		return nil, nil, mongoerrors.New(mongoerrors.ErrOperationFailed, "Persistent cursor connection limit reached; close cursors and retry")
	}
}
