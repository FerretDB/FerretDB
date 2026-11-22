//go:build linux

package clientconn

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

// All watched sockets share one blocking epoll waiter and one shutdown eventfd.
// RDHUP observes FIN without consuming bytes or waking for queued protocol data.
// No timer, per-connection polling or per-connection OS thread is needed.
var peerDisconnects struct {
	sync.Mutex
	poller *peerDisconnectPoller
}

type peerDisconnectWatch struct {
	raw    syscall.RawConn
	cancel context.CancelCauseFunc
	done   chan struct{}
}

type peerDisconnectPoller struct {
	epfd     int
	wakefd   int
	next     uint64
	watches  map[uint64]*peerDisconnectWatch
	stopping bool
	done     chan struct{}
}

func disconnectEvent(id uint64, events uint32) unix.EpollEvent {
	return unix.EpollEvent{Events: events, Fd: int32(id), Pad: int32(id >> 32)}
}

func newPeerDisconnectPoller() (*peerDisconnectPoller, error) {
	epfd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}
	wakefd, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		_ = unix.Close(epfd)
		return nil, err
	}
	event := disconnectEvent(0, unix.EPOLLIN)
	if err = unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, wakefd, &event); err != nil {
		_ = unix.Close(wakefd)
		_ = unix.Close(epfd)
		return nil, err
	}
	p := &peerDisconnectPoller{epfd: epfd, wakefd: wakefd, watches: make(map[uint64]*peerDisconnectWatch), done: make(chan struct{})}
	go p.run()
	return p, nil
}

// stopEmptyLocked detaches an empty poller before waking it. A new connection
// may create its successor while the old waiter releases its descriptors.
func (p *peerDisconnectPoller) stopEmptyLocked() {
	if len(p.watches) != 0 || p.stopping {
		return
	}
	p.stopping = true
	if peerDisconnects.poller == p {
		peerDisconnects.poller = nil
	}
	var value [8]byte
	binary.NativeEndian.PutUint64(value[:], 1)
	// A full eventfd is already readable. EINTR is retried so shutdown cannot
	// strand the waiter or the last connection's cleanup.
	for {
		_, err := unix.Write(p.wakefd, value[:])
		if err != unix.EINTR {
			break
		}
	}
}

func (p *peerDisconnectPoller) remove(id uint64) {
	peerDisconnects.Lock()
	watch := p.watches[id]
	if watch == nil {
		peerDisconnects.Unlock()
		return
	}
	delete(p.watches, id)
	_ = watch.raw.Control(func(fd uintptr) {
		_ = unix.EpollCtl(p.epfd, unix.EPOLL_CTL_DEL, int(fd), nil)
	})
	p.stopEmptyLocked()
	last := len(p.watches) == 0
	peerDisconnects.Unlock()
	if last {
		<-p.done
	}
	close(watch.done)
}

func (p *peerDisconnectPoller) run() {
	defer func() {
		_ = unix.Close(p.wakefd)
		_ = unix.Close(p.epfd)
		close(p.done)
	}()
	events := make([]unix.EpollEvent, 64)
	for {
		n, err := unix.EpollWait(p.epfd, events, -1)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			peerDisconnects.Lock()
			p.stopping = true
			if peerDisconnects.poller == p {
				peerDisconnects.poller = nil
			}
			cancels := make([]context.CancelCauseFunc, 0, len(p.watches))
			for _, watch := range p.watches {
				cancels = append(cancels, watch.cancel)
			}
			peerDisconnects.Unlock()
			for _, cancel := range cancels {
				cancel(err)
			}
			return
		}
		for _, event := range events[:n] {
			peerDisconnects.Lock()
			if p.stopping {
				peerDisconnects.Unlock()
				return
			}
			id := uint64(uint32(event.Fd)) | uint64(uint32(event.Pad))<<32
			watch := p.watches[id]
			peerDisconnects.Unlock()
			if watch != nil {
				watch.cancel(io.EOF)
			}
		}
	}
}

// watchPeerDisconnect also detects FIN while the bounded request reader is
// waiting to hand a pipelined message to a handler blocked in PostgreSQL.
func watchPeerDisconnect(ctx context.Context, conn net.Conn, cancel context.CancelCauseFunc) <-chan struct{} {
	done := make(chan struct{})
	for depth := 0; depth < 4; depth++ {
		wrapped, ok := conn.(interface{ NetConn() net.Conn })
		if !ok {
			break
		}
		conn = wrapped.NetConn()
	}
	tcp, ok := conn.(*net.TCPConn)
	if !ok {
		close(done)
		return done
	}
	raw, err := tcp.SyscallConn()
	if err != nil {
		cancel(err)
		close(done)
		return done
	}
	peerDisconnects.Lock()
	p := peerDisconnects.poller
	if p == nil {
		p, err = newPeerDisconnectPoller()
		if err != nil {
			peerDisconnects.Unlock()
			cancel(err)
			close(done)
			return done
		}
		peerDisconnects.poller = p
	}
	p.next++
	id := p.next
	var addErr error
	err = raw.Control(func(fd uintptr) {
		event := disconnectEvent(id, unix.EPOLLRDHUP|unix.EPOLLHUP|unix.EPOLLERR|unix.EPOLLONESHOT)
		addErr = unix.EpollCtl(p.epfd, unix.EPOLL_CTL_ADD, int(fd), &event)
	})
	if err == nil {
		err = addErr
	}
	if err != nil {
		p.stopEmptyLocked()
		empty := len(p.watches) == 0
		peerDisconnects.Unlock()
		if empty {
			<-p.done
		}
		// Fail closed rather than run a query with a silently missing watcher.
		cancel(err)
		close(done)
		return done
	}
	p.watches[id] = &peerDisconnectWatch{raw: raw, cancel: cancel, done: done}
	peerDisconnects.Unlock()
	context.AfterFunc(ctx, func() { p.remove(id) })
	return done
}
