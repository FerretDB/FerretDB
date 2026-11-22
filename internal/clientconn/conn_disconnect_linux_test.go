//go:build linux

package clientconn

import (
	"bufio"
	"context"
	"net"
	"testing"
)

func TestPipelinedTCPDisconnectCancelsBlockedConsumer(t *testing.T) {
	t.Parallel()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	client, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	server, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	watcher := watchPeerDisconnect(ctx, server, cancel)
	requests, readerDone := readRequestStream(ctx, bufio.NewReader(server), cancel)
	// Deliver one request, then leave another complete request waiting in the
	// bounded reader while its consumer represents a blocked database handler.
	if _, err = client.Write(append(pingMessage(1), pingMessage(2)...)); err != nil {
		t.Fatal(err)
	}
	if req := <-requests; req == nil || req.WireHeader().RequestID != 1 {
		t.Fatal("first request missing")
	}
	if err = client.Close(); err != nil {
		t.Fatal(err)
	}
	awaitClosed(t, ctx.Done())
	awaitClosed(t, watcher)
	awaitClosed(t, readerDone)
}
