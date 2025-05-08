package grpc

import "fmt"

var (
	ErrCannotListenTcpOn = func(loc string, port int, err error) error {
		return fmt.Errorf("cannot listen tcp on %d at %s: %v", port, loc, err)
	}

	ErrCannotServeGRPC = func(loc string, err error) error {
		return fmt.Errorf("cannot serve grpc at %s: %v", loc, err)
	}

	SErrInvalidKey = func(key string) string {
		return fmt.Sprintf("invalid key %s", key)
	}

	SErrFailedToSendMsg = func(msg string) string {
		return fmt.Sprintf("failed to send message: %s", msg)
	}
)
