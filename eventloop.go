package hio

import (
	"log"
	"syscall"
)

type EventLoop struct {
	idx            int
	srv            *Server
	connMap        map[int]*Conn
	newConnections []*Conn
	network        Network
}

func (t *EventLoop) accept(id int, fd int, addr syscall.Sockaddr) {
	err := syscall.SetNonblock(fd, true)
	if err != nil {
		log.Println(err)
		return
	}

	conn := newConn(t, id, fd, addr)
	log.Printf("New connection %s bind to eventloop %v", conn, t.idx)
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
		err := t.network.AddEvents(t.newConnections)
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
				t.closeConn(conn)
			}
			if err := t.writeConn(conn); err != nil {
				t.closeConn(conn)
			}
		}
	}
}

func (t *EventLoop) readConn(conn *Conn) error {
	log.Printf("连接读取事件")

	n, err := syscall.Read(conn.fd, conn.in)
	if err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return err
	}

	if n > 0 {
		conn.Write(conn.in[0:n])
		conn.Close()
	}

	return nil
}

func (t *EventLoop) writeConn(conn *Conn) error {
	if len(conn.out) == 0 {
		return nil
	}

	log.Printf("%v 个字节需要写入", len(conn.out))

	n, err := syscall.Write(conn.fd, conn.out)
	if err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return err
	}

	if n > 0 {
		conn.out = conn.out[n:]
	}

	if conn.state == ConnState_Close {
		t.tryCloseConn(conn)
	}

	return nil
}

func (t *EventLoop) tryCloseConn(conn *Conn) {
	if len(conn.out) > 0 {
		return
	}

	t.closeConn(conn)
}

func (t *EventLoop) closeConn(conn *Conn) {
	err := syscall.Close(conn.fd)
	if err != nil {
		log.Printf("close connection failed: %v", err)
	}
	delete(t.connMap, conn.fd)
}

func (t *EventLoop) connCount() int {
	return len(t.connMap)
}

func newEventLoop(srv *Server, idx int) *EventLoop {
	loop := &EventLoop{}
	loop.idx = idx
	loop.srv = srv
	loop.connMap = make(map[int]*Conn)
	loop.newConnections = make([]*Conn, 0)

	network, err := newNetwork()
	if err != nil {
		panic(err)
	}
	loop.network = network

	return loop
}
