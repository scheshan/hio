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

func (t *EventLoop) accept(fd int, addr syscall.Sockaddr) {
	err := syscall.SetNonblock(fd, true)
	if err != nil {
		log.Println(err)
		return
	}

	conn := newConn(t, fd, addr)
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
	fds, n, err := t.network.Wait(5000)
	if err != nil {
		panic(err)
	}

	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		fd := fds[i]
		conn := t.connMap[fd]
		if conn != nil {
			t.readConn(conn)
			t.writeConn(conn)
		}
	}
}

func (t *EventLoop) readConn(conn *Conn) {
	log.Printf("连接读取事件")

	n, err := syscall.Read(conn.fd, conn.in)
	if err != nil {
		if err != syscall.EAGAIN {
			t.closeConn(conn)
		}
		return
	}

	if n > 0 {
		conn.Write(conn.in[0:n])
	}
}

func (t *EventLoop) writeConn(conn *Conn) {
	if len(conn.out) == 0 {
		return
	}

	log.Printf("%v 个字节需要写入", len(conn.out))

	n, err := syscall.Write(conn.fd, conn.out)
	if err != nil {
		if err != syscall.EAGAIN {
			t.closeConn(conn)
		}
		return
	}

	if n > 0 {
		conn.out = conn.out[n:]
	}
}

func (t *EventLoop) closeConn(conn *Conn) {
	syscall.Close(conn.fd)
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
