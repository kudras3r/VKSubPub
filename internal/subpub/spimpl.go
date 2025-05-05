package subpub

/*

	there are subpub struct looks like:
	subject1 : [
		subscriber1 [
			sub1 queue [ msg1, msg2, ... ]
			sub1 extraQueue [ exmsg1, exmsg2, ... ]
			sub1 handler()
		],
		subscriber2 [
			sub2 queue [ msg1, msg2, ... ]
			...
		],
		...
	]
	subject2 : [ ... ]

*/

import (
	"context"
	"sync"
)

// ! TODO : in future set as config values
const (
	GLOC = "internal/subpub/subimpl.go/" // for logging

	QSIZE          = 8
	MAX_EXTRA_SIZE = 4056 // if there - sub is REALLY slow

	DEFAULT_SUBS_CAP  = 64
	DEFAULT_EXTRA_CAP = 16
)

type subscriber struct {
	ch   chan interface{}
	hl   MessageHandler
	stop chan struct{}

	extra []interface{}
	exMu  sync.Mutex
}

// For 1 write : M reads (IO bound)
type subPub struct {
	mu     sync.RWMutex
	subcrs map[string][]*subscriber

	wg sync.WaitGroup
}

func emptySubject(subject string) bool {
	return subject == ""
}

func nilHandler(hl MessageHandler) bool {
	return hl == nil
}

func (s *subscriber) drainExtra() {
	s.exMu.Lock()
	defer s.exMu.Unlock()

	if len(s.extra) > MAX_EXTRA_SIZE {
		// ? THINK : what we gonna do
	}

	for len(s.extra) > 0 {
		select {
		case s.ch <- s.extra[0]:
			s.extra = s.extra[1:]
		default:
			return
		}
	}
}

// Subscribe creates an asynchronous queue subscriber on the given subject.
func (s *subPub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	loc := GLOC + "subPub.Subscribe()"

	if emptySubject(subject) {
		return nil, ErrEmptySubject(loc)
	}
	if nilHandler(cb) {
		return nil, ErrEmptyHandler(loc)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subcrs[subject]; !ok {
		s.subcrs[subject] = make([]*subscriber, 0, DEFAULT_SUBS_CAP)
	}

	nsub := &subscriber{
		ch:   make(chan interface{}, QSIZE),
		hl:   cb,
		stop: make(chan struct{}),

		extra: make([]interface{}, 0, DEFAULT_EXTRA_CAP),
	}
	s.subcrs[subject] = append(s.subcrs[subject], nsub)

	s.wg.Add(1)

	// listen
	go func() {
		defer s.wg.Done()
		for {
			select {
			case msg := <-nsub.ch:
				// ? THINK : may be recover
				nsub.hl(msg) // if hl panic - goroutine will be dead
				nsub.drainExtra()
			case <-nsub.stop:
				return
			}
		}
	}()

	return &subscription{
		subject: subject,
		sub:     nsub,
		sp:      s,
	}, nil
}

func (s *subPub) Publish(subject string, msg interface{}) error {
	loc := GLOC + "subPub.Publish()"

	if emptySubject(subject) {
		return ErrEmptySubject(loc)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	subs, ok := s.subcrs[subject]
	if !ok {
		return nil
	}

	for _, sub := range subs {
		// ! TODO : Один медленный подписчик не должен тормозить остальных (+)
		select {
		case sub.ch <- msg:
		default:
			// if is full
			sub.exMu.Lock()
			sub.extra = append(sub.extra, msg)
			sub.exMu.Unlock()
		}
	}

	return nil
}

func (s *subPub) Close(ctx context.Context) error {
	loc := GLOC + "Close()"

	_ = loc

	s.mu.Lock()
	for _, subs := range s.subcrs {
		for _, sub := range subs {
			close(sub.stop)
		}
	}
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
