package hio

type ServerOptions struct {
	LoopNum     uint32
	Port        int
	LoadBalance string
	Listener    *Listener
}
