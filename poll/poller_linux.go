package poll

import "golang.org/x/sys/unix"

var wakeupData = []byte{1, 0, 0, 0, 0, 0, 0, 0}

type Poller struct {
	ep      int
	ef      int
	eEvents []unix.EpollEvent
	pEvents []PollerEvent
}

func (t *Poller) Add(fd int) error {
	return unix.EpollCtl(t.ep, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Fd:     int32(fd),
		Events: unix.EPOLLIN,
	})
}

func (t *Poller) Delete(fd int) error {
	return unix.EpollCtl(t.ep, unix.EPOLL_CTL_DEL, fd, nil)
}

func (t *Poller) EnableWrite(fd int) error {
	return unix.EpollCtl(t.ep, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{
		Fd:     int32(fd),
		Events: unix.EPOLLIN | unix.EPOLLOUT,
	})
}

func (t *Poller) DisableWrite(fd int) error {
	return unix.EpollCtl(t.ep, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{
		Fd:     int32(fd),
		Events: unix.EPOLLIN,
	})
}

func (t *Poller) Wait(timeoutMs int64) ([]PollerEvent, error) {
	n, err := unix.EpollWait(t.ep, t.eEvents, int(timeoutMs))
	if err != nil {
		return nil, err
	}

	for i := 0; i < n; i++ {
		ee := t.eEvents[i]
		pe := t.pEvents[i]

		pe.id = int(ee.Fd)
		pe.typ = 0

		if ee.Events&unix.EPOLLIN > 0 {
			pe.typ |= 1
		}
		if ee.Events&unix.EPOLLOUT > 0 {
			pe.typ |= 2
		}
	}

	res := t.pEvents[:n]

	if n == len(t.pEvents) {
		t.incrEvents(n << 1)
	}

	return res, nil
}

func (t *Poller) Wakeup() error {
	_, err := unix.Write(t.ef, wakeupData)
	return err
}

func (t *Poller) Close() {
	if t.ep > 0 {
		unix.Close(t.ep)
		t.ep = 0
	}
	if t.ef > 0 {
		unix.Close(t.ef)
		t.ef = 0
	}

	t.eEvents = nil
	t.pEvents = nil
}

func (t *Poller) incrEvents(size int) {
	t.eEvents = make([]unix.EpollEvent, size)
	t.pEvents = make([]PollerEvent, size)
}

func NewPoller() (*Poller, error) {
	ep, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	p := &Poller{}
	p.ep = ep

	ef, err := unix.Eventfd(0, 0)
	if err != nil {
		p.Close()
		return nil, err
	}
	p.ef = ef

	err = p.Add(ef)
	if err != nil {
		p.Close()
		return nil, err
	}

	p.incrEvents(1024)

	return p, nil
}
