// +build darwin

package hio

import (
	"syscall"
)

type network struct {
	kq      int
	changes []syscall.Kevent_t
	events  []syscall.Kevent_t
	timeout *syscall.Timespec
	fds     []int
}

func (t *network) AddEvents(conns ...*Conn) error {
	for _, conn := range conns {
		t.changes = append(
			t.changes,
			syscall.Kevent_t{
				Ident: uint64(conn.fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ | syscall.EVFILT_WRITE,
			})
	}

	return nil
}

func (t *network) RemoveEvents(conns ...*Conn) error {
	for _, conn := range conns {
		t.changes = append(
			t.changes,
			syscall.Kevent_t{
				Ident: uint64(conn.fd), Flags: syscall.EV_DELETE, Filter: syscall.EVFILT_READ | syscall.EVFILT_WRITE,
			})
	}

	return nil
}

func (t *network) Wait(timeMs int) (fds []int, n int, err error) {
	if timeMs > 1000 {
		t.timeout.Sec = int64(timeMs / 1000)
	}
	t.timeout.Nsec = int64(timeMs%1000) * 100000

	n, err = syscall.Kevent(t.kq, t.changes, t.events, t.timeout)
	if err != nil {
		return nil, 0, err
	}

	t.changes = t.changes[:0]

	if n > 0 {
		for i := 0; i < n; i++ {
			t.fds[i] = int(t.events[i].Ident)
		}
	}
	return t.fds, n, nil
}

func newNetwork() (*network, error) {
	kq, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}

	eventsNum := 1024

	n := &network{}
	n.kq = kq
	n.changes = make([]syscall.Kevent_t, 0)

	n.events = make([]syscall.Kevent_t, eventsNum, eventsNum)
	n.fds = make([]int, eventsNum, eventsNum)
	for i := 0; i < eventsNum; i++ {
		n.events = append(n.events, syscall.Kevent_t{})
		n.fds = append(n.fds, 0)
	}

	n.timeout = &syscall.Timespec{}

	return n, nil
}
