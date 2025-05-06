package grpc

import "fmt"

var (
	ErrCannotListenTcpOn = func(loc string, port int, err error) error {
		return fmt.Errorf("cannot listen tcp on %d at %s: %v", port, loc, err)
	}

	ErrCannotServeGRPC = func(loc string, err error) error {
		return fmt.Errorf("cannot serve grpc at %s: %v", loc, err)
	}
)
