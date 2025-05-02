package subpub

import "fmt"

var (
	ErrEmptySubject = func(loc string) error { return fmt.Errorf("error at %s: empty subject", loc) }
	ErrEmptyHandler = func(loc string) error { return fmt.Errorf("error at %s: empty handler", loc) }
)
