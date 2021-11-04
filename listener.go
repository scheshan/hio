package hio

type Listener struct {
	OnConnOpened func(conn *Conn)
	OnConnClosed func(conn *Conn, e error)
	OnConnRead   func(conn *Conn)
}
