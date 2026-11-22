package clientconn

import (
	"context"
	"encoding/binary"
	"io"
	"testing"
	"time"
)

func pingMessage(id uint32) []byte {
	doc := make([]byte, 4)
	doc = append(doc, 0x10, 'p', 'i', 'n', 'g', 0, 1, 0, 0, 0)
	doc = append(doc, 0x02, '$', 'd', 'b', 0, 6, 0, 0, 0, 'a', 'd', 'm', 'i', 'n', 0, 0)
	binary.LittleEndian.PutUint32(doc, uint32(len(doc)))
	message := make([]byte, 21)
	message = append(message, doc...)
	binary.LittleEndian.PutUint32(message, uint32(len(message)))
	binary.LittleEndian.PutUint32(message[4:], id)
	binary.LittleEndian.PutUint32(message[12:], 2013)
	return message
}

func awaitClosed(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("request reader did not finish")
	}
}

func TestRequestStreamDisconnectCancelsActiveRequest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	reader, writer := io.Pipe()
	defer reader.Close()
	requests, done := readRequestStream(ctx, reader, cancel)
	if _, err := writer.Write(pingMessage(1)); err != nil {
		t.Fatal(err)
	}
	if req := <-requests; req == nil || req.Document().Command() != "ping" {
		t.Fatal("valid request was not delivered")
	}
	// The consumer is executing that request and does not read another one.
	_ = writer.Close()
	awaitClosed(t, ctx.Done())
	awaitClosed(t, done)
	if context.Cause(ctx) == nil {
		t.Fatal("disconnect did not provide a cancellation cause")
	}
}

func TestRequestStreamPreservesPipelinedOrder(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	reader, writer := io.Pipe()
	defer reader.Close()
	requests, done := readRequestStream(ctx, reader, cancel)
	go func() {
		defer writer.Close()
		for id := uint32(1); id <= 3; id++ {
			if _, err := writer.Write(pingMessage(id)); err != nil {
				return
			}
		}
	}()
	for id := int32(1); id <= 3; id++ {
		select {
		case req := <-requests:
			if req == nil || req.WireHeader().RequestID != id {
				t.Fatal("pipelined request order changed")
			}
		case <-time.After(time.Second):
			t.Fatal("pipelined request was not delivered")
		}
	}
	awaitClosed(t, done)
}

func TestRequestStreamCancellationUnblocksQueuedMessage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancelCause(context.Background())
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	_, done := readRequestStream(ctx, reader, cancel)
	if _, err := writer.Write(pingMessage(1)); err != nil {
		t.Fatal(err)
	}
	cancel(context.Canceled)
	awaitClosed(t, done)
}
