package subpub

/*

	there are subpub struct looks like:
	subject1 : [
		subscriber1 [
			sub1 queue [ msg1, msg2, ... ]
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

	QSIZE            = 64
	DEFAULT_SUBS_CAP = 64
)

type subscriber struct {
	ch   chan interface{}
	hl   MessageHandler
	stop chan struct{}
}

// For 1 write : M reads (IO bound)
type subPub struct {
	mu     sync.RWMutex
	subcrs map[string][]*subscriber
}

func emptySubject(subject string) bool {
	return subject == ""
}

// Subscribe creates an asynchronous queue subscriber on the given subject.
func (s *subPub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	loc := GLOC + "subPub.Subscribe()"

	if emptySubject(subject) {
		return nil, ErrEmptySubject(loc)
	}
	if cb == nil {
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
	}
	s.subcrs[subject] = append(s.subcrs[subject], nsub)

	// listen
	go func() {
		for {
			select {
			case msg := <-nsub.ch:
				nsub.hl(msg)
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
		// ! TODO : fix blocking if sub.ch is full
		// ! Один медленный подписчик не должен тормозить остальных
		sub.ch <- msg
	}

	return nil
}

func (s *subPub) Close(ctx context.Context) error {
	panic("Implement me")
}
