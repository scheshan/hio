package hio

import "syscall"

type network struct {
	ep           int
	nwEvents     []networkEvent
	events       []syscall.EpollEvent
	lastTimeout  int64
	lastTimespec *syscall.Timespec
}

func (t *network) addRead(fd int) error {
	return t.addEvent(fd, syscall.EPOLL_CTL_ADD, syscall.EPOLLIN)
}

func (t *network) addReadWrite(fd int) error {
	return t.addEvent(fd, syscall.EPOLL_CTL_ADD, syscall.EPOLLIN|syscall.EPOLLOUT)
}

func (t *network) addWrite(fd int) error {
	return t.addEvent(fd, syscall.EPOLL_CTL_ADD, syscall.EPOLLOUT)
}

func (t *network) removeRead(fd int) error {
	return t.addEvent(fd, syscall.EPOLL_CTL_DEL, syscall.EPOLLIN)
}

func (t *network) removeReadWrite(fd int) error {
	return t.addEvent(fd, syscall.EPOLL_CTL_DEL, syscall.EPOLLIN|syscall.EPOLLOUT)
}

func (t *network) removeWrite(fd int) error {
	return t.addEvent(fd, syscall.EPOLL_CTL_DEL, syscall.EPOLLOUT)
}

func (t *network) wait(timeoutMs int64) (events []networkEvent, err error) {
	if t.lastTimeout != timeoutMs {
		t.lastTimeout = timeoutMs

		ts := syscall.NsecToTimespec(timeoutMs * 1000000)
		t.lastTimespec = &ts
	}

	n, err := syscall.EpollWait(t.ep, t.events, int(timeoutMs))
	if err != nil {
		return nil, err
	}

	for i := 0; i < n; i++ {
		ev := t.events[i]
		t.nwEvents[i].fd = int(ev.Fd)
		t.nwEvents[i].ev = 0

		if ev.Events&syscall.EPOLLIN > 0 {
			t.nwEvents[i].ev |= 1
		}
		if ev.Events&syscall.EPOLLOUT > 0 {
			t.nwEvents[i].ev |= 2
		}
	}

	return t.nwEvents[:n], nil
}

func (t *network) addEvent(fd int, mode int, events uint32) error {
	event := &syscall.EpollEvent{}
	event.Fd = int32(fd)
	event.Events = events

	return syscall.EpollCtl(t.ep, fd, mode, event)
}

func (t *network) shutdown() {
	syscall.Close(t.ep)
}

func newNetwork() (*network, error) {
	fd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	evSize := 1024

	nw := new(network)
	nw.ep = fd
	nw.nwEvents = make([]networkEvent, evSize, evSize)
	nw.events = make([]syscall.EpollEvent, evSize, evSize)

	return nw, nil
}
