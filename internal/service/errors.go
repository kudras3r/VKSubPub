package service

import "fmt"

var (
	ErrFailedToSubscribe = func(key string) error {
		return fmt.Errorf("failed to subscribe at %s", key)
	}

	ErrInvalidKey = func(key string) error {
		return fmt.Errorf("invalid key %s", key)
	}
)
