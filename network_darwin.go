package hio

import "syscall"

type network struct {
	kq           int
	nwEvents     []networkEvent
	events       []syscall.Kevent_t
	lastTimeout  int64
	lastTimespec *syscall.Timespec
}

func (t *network) addRead(fd int) error {
	return t.addEvent(fd, syscall.EVFILT_READ, syscall.EV_ADD)
}

func (t *network) addReadWrite(fd int) error {
	return t.addEvent(fd, syscall.EVFILT_READ|syscall.EVFILT_WRITE, syscall.EV_ADD)
}

func (t *network) addWrite(fd int) error {
	return t.addEvent(fd, syscall.EVFILT_WRITE, syscall.EV_ADD)
}

func (t *network) removeRead(fd int) error {
	return t.addEvent(fd, syscall.EVFILT_READ, syscall.EV_DELETE)
}

func (t *network) removeReadWrite(fd int) error {
	return t.addEvent(fd, syscall.EVFILT_READ|syscall.EVFILT_WRITE, syscall.EV_DELETE)
}

func (t *network) removeWrite(fd int) error {
	return t.addEvent(fd, syscall.EVFILT_WRITE, syscall.EV_DELETE)
}

func (t *network) wait(timeoutMs int64) (events []networkEvent, err error) {
	if t.lastTimeout != timeoutMs {
		t.lastTimeout = timeoutMs

		ts := syscall.NsecToTimespec(timeoutMs * 1000000)
		t.lastTimespec = &ts
	}

	n, err := syscall.Kevent(t.kq, nil, t.events, t.lastTimespec)
	if err != nil {
		return nil, err
	}

	for i := 0; i < n; i++ {
		ev := t.events[i]
		t.nwEvents[i].fd = int(ev.Ident)
		t.nwEvents[i].ev = 0

		if ev.Filter|syscall.EVFILT_READ == syscall.EVFILT_READ {
			t.nwEvents[i].ev |= 1
		}
		if ev.Filter|syscall.EVFILT_WRITE == syscall.EVFILT_WRITE {
			t.nwEvents[i].ev |= 2
		}
	}

	return t.nwEvents[:n], nil
}

func (t *network) addEvent(fd int, filter int16, flags uint16) error {
	changes := make([]syscall.Kevent_t, 1, 1)
	changes[0].Ident = uint64(fd)
	changes[0].Filter = filter
	changes[0].Flags = uint16(flags)

	_, err := syscall.Kevent(t.kq, changes, nil, nil)
	return err
}

func (t *network) shutdown() {
	syscall.Close(t.kq)
}

func newNetwork() (*network, error) {
	fd, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}

	evSize := 1024

	nw := new(network)
	nw.kq = fd
	nw.nwEvents = make([]networkEvent, evSize, evSize)
	nw.events = make([]syscall.Kevent_t, evSize, evSize)

	return nw, nil
}
