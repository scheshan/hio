package poll

import "golang.org/x/sys/unix"

type Poller struct {
	kq      int
	kEvents []unix.Kevent_t
	pEvents []PollerEvent
	ts      *unix.Timespec
}

func (t *Poller) AddRead(fd int) error {
	return t.addChanges(fd, unix.EV_ADD, unix.EVFILT_READ)
}

func (t *Poller) AddWrite(fd int) error {
	return t.addChanges(fd, unix.EV_ADD, unix.EVFILT_WRITE)
}

func (t *Poller) AddReadWrite(fd int) error {
	return t.addChanges(fd, unix.EV_ADD, unix.EVFILT_READ, unix.EVFILT_WRITE)
}

func (t *Poller) RemoveRead(fd int) error {
	return t.addChanges(fd, unix.EV_DELETE, unix.EVFILT_READ)
}

func (t *Poller) RemoveWrite(fd int) error {
	return t.addChanges(fd, unix.EV_DELETE, unix.EVFILT_WRITE)
}

func (t *Poller) RemoveReadWrite(fd int) error {
	return t.addChanges(fd, unix.EV_DELETE, unix.EVFILT_READ, unix.EVFILT_WRITE)
}

func (t *Poller) Wait(timeoutMs int64) ([]PollerEvent, error) {
	t.ts.Sec = timeoutMs / 1000
	t.ts.Nsec = (timeoutMs % 1000) * 1000000

	n, err := unix.Kevent(t.kq, nil, t.kEvents, t.ts)
	if err != nil {
		return nil, err
	}

	for i := 0; i < n; i++ {
		ke := t.kEvents[i]

		t.pEvents[i].id = int(ke.Ident)
		t.pEvents[i].typ = 0

		if ke.Filter == unix.EVFILT_READ {
			t.pEvents[i].typ |= 1
		}
		if ke.Filter == unix.EVFILT_WRITE {
			t.pEvents[i].typ |= 2
		}
	}

	res := t.pEvents[:n]

	if n == len(t.pEvents) {
		t.incrEvents(n << 1)
	}

	return res, nil
}

func (t *Poller) Wakeup() error {
	_, err := unix.Kevent(t.kq, []unix.Kevent_t{{
		Ident:  0,
		Filter: unix.EVFILT_USER,
		Fflags: unix.NOTE_TRIGGER,
	}}, nil, nil)
	return err
}

func (t *Poller) Close() {
	unix.Close(t.kq)
}

func (t *Poller) addChanges(fd int, flags uint16, filters ...int16) error {
	changes := make([]unix.Kevent_t, len(filters))

	for i, filter := range filters {
		changes[i].Ident = uint64(fd)
		changes[i].Filter = filter
		changes[i].Flags = flags
	}

	_, err := unix.Kevent(t.kq, changes, nil, nil)
	return err
}

func (t *Poller) incrEvents(size int) {
	t.kEvents = make([]unix.Kevent_t, size)
	t.pEvents = make([]PollerEvent, size)
}

func NewPoller() (*Poller, error) {
	kq, err := unix.Kqueue()
	if err != nil {
		return nil, err
	}

	p := &Poller{}
	p.kq = kq

	err = p.addChanges(0, unix.EV_ADD|unix.EV_CLEAR, unix.EVFILT_USER)
	if err != nil {
		p.Close()
		return nil, err
	}

	p.ts = &unix.Timespec{}
	p.incrEvents(1024)

	return p, nil
}
