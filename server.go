package hio

import "errors"

type Server interface {
	Serve() error
	Shutdown()
}

type mockServer struct {
}

func (t *mockServer) Serve() error {
	return errors.New("unsupported")
}

func (t *mockServer) Shutdown() {

}
