package conninfo

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnInfo(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		peerAddr net.Addr
	}{
		"EmptyPeerAddr": {
			peerAddr: nil,
		},
		"NonEmptyPeerAddr": {
			peerAddr: &net.TCPAddr{
				IP:   net.IPv4(127, 0, 0, 8),
				Port: 1234,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			connInfo := &ConnInfo{
				PeerAddr: tc.peerAddr,
			}
			ctx = WithConnInfo(ctx, connInfo)
			actual := GetConnInfo(ctx)
			assert.Equal(t, *connInfo, *actual)
		})
	}
}
