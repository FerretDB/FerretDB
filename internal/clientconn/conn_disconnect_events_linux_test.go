//go:build linux

package clientconn

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func disconnectTCPPair(t *testing.T) (*net.TCPConn, *net.TCPConn) {
	t.Helper()
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	client, err := net.DialTCP("tcp", nil, listener.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatal(err)
	}
	server, err := listener.AcceptTCP()
	if err != nil {
		client.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close(); server.Close() })
	return client, server
}

func disconnectWatch(t *testing.T, server net.Conn) (context.Context, context.CancelCauseFunc, <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithCancelCause(context.Background())
	done := watchPeerDisconnect(ctx, server, cancel)
	t.Cleanup(func() { cancel(context.Canceled); awaitClosed(t, done) })
	return ctx, cancel, done
}

func TestDisconnectWatcherPreservesQueuedProtocolBytes(t *testing.T) {
	client, server := disconnectTCPPair(t)
	ctx, _, _ := disconnectWatch(t, server)
	wanted := append(pingMessage(1), pingMessage(2)...)
	if _, err := client.Write(wanted); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ctx.Done():
		t.Fatal("readable protocol data was treated as a disconnect")
	case <-time.After(50 * time.Millisecond):
	}
	got := make([]byte, len(wanted))
	server.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := io.ReadFull(server, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, wanted) {
		t.Fatal("watcher consumed or changed protocol bytes")
	}
}

func TestDisconnectWatcherHalfCloseAndReset(t *testing.T) {
	for _, reset := range []bool{false, true} {
		name := "FIN"
		if reset {
			name = "RST"
		}
		t.Run(name, func(t *testing.T) {
			client, server := disconnectTCPPair(t)
			ctx, _, done := disconnectWatch(t, server)
			var err error
			if reset {
				client.SetLinger(0)
				err = client.Close()
			} else {
				err = client.CloseWrite()
			}
			if err != nil {
				t.Fatal(err)
			}
			awaitClosed(t, ctx.Done())
			awaitClosed(t, done)
			if !errors.Is(context.Cause(ctx), io.EOF) {
				t.Fatalf("unexpected cause: %v", context.Cause(ctx))
			}
		})
	}
}

func TestDisconnectWatcherServerCancellationKeepsSocketOwnedByCaller(t *testing.T) {
	client, server := disconnectTCPPair(t)
	ctx, cancel, done := disconnectWatch(t, server)
	wanted := errors.New("server shutdown")
	cancel(wanted)
	awaitClosed(t, done)
	if context.Cause(ctx) != wanted {
		t.Fatal("cancellation cause changed")
	}
	if _, err := client.Write([]byte{42}); err != nil {
		t.Fatal(err)
	}
	server.SetReadDeadline(time.Now().Add(time.Second))
	var data [1]byte
	if _, err := io.ReadFull(server, data[:]); err != nil || data[0] != 42 {
		t.Fatal("watcher closed the caller's socket", err)
	}
}

func TestDisconnectWatcherSharedLifecycleAndDescriptorReuse(t *testing.T) {
	_, anchor := disconnectTCPPair(t)
	ctx, cancel, done := disconnectWatch(t, anchor)
	peerDisconnects.Lock()
	p := peerDisconnects.poller
	peerDisconnects.Unlock()
	if p == nil {
		t.Fatal("no shared waiter")
	}
	for i := 0; i < 64; i++ {
		client, server := disconnectTCPPair(t)
		other, _, stopped := disconnectWatch(t, server)
		peerDisconnects.Lock()
		shared := peerDisconnects.poller == p
		peerDisconnects.Unlock()
		if !shared {
			t.Fatal("connections created separate waiters")
		}
		client.Close()
		awaitClosed(t, other.Done())
		awaitClosed(t, stopped)
		server.Close()
		select {
		case <-ctx.Done():
			t.Fatal("stale descriptor event canceled another socket")
		default:
		}
	}
	cancel(context.Canceled)
	awaitClosed(t, done)
	awaitClosed(t, p.done)
	peerDisconnects.Lock()
	empty := peerDisconnects.poller == nil && len(p.watches) == 0
	peerDisconnects.Unlock()
	if !empty {
		t.Fatal("last connection left a waiter or registration behind")
	}
}

func TestDisconnectWatcherRegistrationFailureCancelsConnection(t *testing.T) {
	_, server := disconnectTCPPair(t)
	server.Close()
	ctx, _, done := disconnectWatch(t, server)
	awaitClosed(t, ctx.Done())
	awaitClosed(t, done)
	if context.Cause(ctx) == nil {
		t.Fatal("failed registration left unprotected work running")
	}
}

func TestDisconnectWatcherConcurrentChurn(t *testing.T) {
	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for n := 0; n < 8; n++ {
				client, server := disconnectTCPPair(t)
				ctx, cancel, done := disconnectWatch(t, server)
				if n%2 == 0 {
					client.Close()
				} else {
					cancel(context.Canceled)
				}
				awaitClosed(t, ctx.Done())
				awaitClosed(t, done)
				client.Close()
				server.Close()
			}
		}()
	}
	workers.Wait()
	peerDisconnects.Lock()
	empty := peerDisconnects.poller == nil
	peerDisconnects.Unlock()
	if !empty {
		t.Fatal("concurrent churn left a shared waiter behind")
	}
}
