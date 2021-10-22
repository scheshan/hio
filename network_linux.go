// +build linux

package hio

import "syscall"

type network struct {
	ep     int
	events []syscall.EpollEvent
	fds    []int
}

func (t *network) AddEvents(conns []*Conn) error {
	for _, conn := range conns {
		if err := syscall.EpollCtl(t.ep, syscall.EPOLL_CTL_ADD, conn.fd, &syscall.EpollEvent{
			Events: syscall.EPOLLIN | syscall.EPOLLOUT,
			Fd:     int32(conn.fd),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (t *network) RemoveEvents(conns []*Conn) error {
	for _, conn := range conns {
		if err := syscall.EpollCtl(t.ep, syscall.EPOLL_CTL_DEL, conn.fd, &syscall.EpollEvent{
			Events: syscall.EPOLLIN | syscall.EPOLLOUT,
			Fd:     int32(conn.fd),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (t *network) Wait(timeMs int) (fds []int, n int, err error) {
	n, err = syscall.EpollWait(t.ep, t.events, timeMs)
	if err != nil {
		return nil, 0, err
	}

	if n > 0 {
		for i := 0; i < n; i++ {
			t.fds[i] = int(t.events[i].Fd)
		}
	}

	return t.fds, n, nil
}

func newNetwork() (*network, error) {
	ep, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	eventNum := 1024

	n := &network{}
	n.ep = ep
	n.fds = make([]int, eventNum, eventNum)

	n.events = make([]syscall.EpollEvent, eventNum, eventNum)
	for i := 0; i < eventNum; i++ {
		n.events[i] = syscall.EpollEvent{}
	}

	return n, nil
}
