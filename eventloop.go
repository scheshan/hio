package hio

import (
	"io"
	"log"
	"syscall"
)

type EventLoop struct {
	idx            int
	srv            *Server
	connMap        map[int]*Conn
	newConnections []*Conn
	network        Network
	bufIn          []byte
	bufOut         []byte
	mp             *MemoryPool
}

func (t *EventLoop) accept(id int, fd int, addr syscall.Sockaddr) {
	err := syscall.SetNonblock(fd, true)
	if err != nil {
		log.Println(err)
		return
	}

	conn := newConn(t, id, fd, addr)
	log.Printf("New connection %s bind to eventloop %v", conn, t.idx)
	t.triggerConnOpen(conn)

	t.newConnections = append(t.newConnections, conn)
}

func (t *EventLoop) loop() {
	for t.srv.isRunning() {
		t.handleNewConnection()
		t.handleIOEvents()
	}
}

func (t *EventLoop) handleNewConnection() {
	if len(t.newConnections) == 0 {
		return
	}

	if len(t.newConnections) > 0 {
		err := t.network.AddEvents(t.newConnections...)
		if err != nil {
			panic(err)
		}
	}

	for _, conn := range t.newConnections {
		t.connMap[conn.fd] = conn
	}

	t.newConnections = t.newConnections[:0]
}

func (t *EventLoop) handleIOEvents() {
	fds, n, err := t.network.Wait(100)
	if err != nil && err != syscall.EINTR {
		panic(err)
	}

	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		fd := fds[i]
		conn := t.connMap[fd]
		if conn != nil {
			if err := t.readConn(conn); err != nil {
				t.closeConn(conn, err)
			}
			if err := t.writeConn(conn); err != nil {
				t.closeConn(conn, err)
			}
		}
	}
}

func (t *EventLoop) readConn(conn *Conn) error {
	log.Printf("连接读取事件")

	n, err := syscall.Read(conn.fd, t.bufIn)
	if n == 0 || err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		if n == 0 {
			return io.EOF
		}
		return err
	}

	conn.in.Write(t.bufIn[:n])

	t.triggerConnRead(conn)
	return nil
}

func (t *EventLoop) writeConn(conn *Conn) error {
	if !conn.out.CanRead() {
		return nil
	}

	n, err := conn.out.CopyToFile(conn.fd)
	if n == 0 {
		t.closeConn(conn, nil)
		return nil
	}
	if err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return err
	}

	if conn.state == ConnState_Close {
		t.tryCloseConn(conn)
	}

	return nil
}

func (t *EventLoop) tryCloseConn(conn *Conn) {
	if conn.out.CanRead() {
		return
	}

	t.closeConn(conn, nil)
}

func (t *EventLoop) closeConn(conn *Conn, e error) {
	if e != nil {
		log.Printf("Close connection %s: %v", conn, e)
	}

	err := syscall.Close(conn.fd)
	if err != nil {
		log.Printf("close connection failed: %v", err)
	}
	delete(t.connMap, conn.fd)
	t.network.RemoveEvents(conn)
	t.triggerConnClose(conn, e)
}

func (t *EventLoop) connCount() int {
	return len(t.connMap)
}

func (t *EventLoop) triggerConnOpen(conn *Conn) {
	if t.srv.listener != nil && t.srv.listener.OnConnOpened != nil {
		t.srv.listener.OnConnOpened(conn)
	}
}

func (t *EventLoop) triggerConnClose(conn *Conn, err error) {
	if t.srv.listener != nil && t.srv.listener.OnConnClosed != nil {
		t.srv.listener.OnConnClosed(conn, err)
	}
}

func (t *EventLoop) triggerConnRead(conn *Conn) {
	if t.srv.listener != nil && t.srv.listener.OnConnRead != nil {
		t.srv.listener.OnConnRead(conn)
	}
}

func newEventLoop(srv *Server, idx int) *EventLoop {
	loop := &EventLoop{}
	loop.idx = idx
	loop.srv = srv
	loop.connMap = make(map[int]*Conn)
	loop.newConnections = make([]*Conn, 0)
	loop.bufIn = make([]byte, 40960)
	loop.bufOut = make([]byte, 40960)
	loop.mp = srv.mp

	network, err := newNetwork()
	if err != nil {
		panic(err)
	}
	loop.network = network

	return loop
}
