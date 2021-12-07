package hio

type EventFunc func()

func connEventFunc(f func(conn *Conn), conn *Conn) EventFunc {
	return func() {
		f(conn)
	}
}
