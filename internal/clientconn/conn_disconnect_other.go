//go:build !linux

package clientconn

import (
	"context"
	"net"
)

// Other platforms retain cancellation through the single request reader.
func watchPeerDisconnect(context.Context, net.Conn, context.CancelCauseFunc) <-chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
}
