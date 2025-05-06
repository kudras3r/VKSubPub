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

const (
	GLOC_SP = "internal/subpub/spimpl.go/" // for logging
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
	loc := GLOC_SP + "drainExtra()"

	s.exMu.Lock()
	defer s.exMu.Unlock()

	if len(s.extra) > conf.MaxExSize {
		log.Warnf("%s: extra slice size=%d when max=%d!", loc, len(s.extra), conf.MaxExSize)
		// ? THINK : what we gonna do
	}

	for len(s.extra) > 0 {
		log.Warnf("%s: have extra messages!", loc)
		select {
		case s.ch <- s.extra[0]:
			s.extra = s.extra[1:]
		default:
			// ch is full
			return
		}
	}
}

// Subscribe creates an asynchronous queue subscriber on the given subject.
func (s *subPub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	loc := GLOC_SP + "Subscribe()"

	if emptySubject(subject) {
		log.Warnf("%s: emtry subject!", loc)
		return nil, ErrEmptySubject(loc)
	}
	if nilHandler(cb) {
		log.Warnf("%s: handler is nil!", loc)
		return nil, ErrEmptyHandler(loc)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subcrs[subject]; !ok {
		s.subcrs[subject] = make([]*subscriber, 0, conf.DefaultSubsCap)
	}

	nsub := &subscriber{
		ch:   make(chan interface{}, conf.SubQSize),
		hl:   cb,
		stop: make(chan struct{}),

		extra: make([]interface{}, 0, conf.DefaultExCap),
	}
	s.subcrs[subject] = append(s.subcrs[subject], nsub)
	log.Infof("%s: new sub added to subject: %s", loc, subject)

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
				log.Infof("%s: stop sub with subject: %s", loc, subject)
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
	loc := GLOC_SP + "Publish()"

	if emptySubject(subject) {
		log.Warnf("%s: emtry subject!", loc)
		return ErrEmptySubject(loc)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	subs, ok := s.subcrs[subject]
	if !ok {
		return nil
	}

	log.Infof("%s: publish { subject: %s, msg: %s }", loc, subject, msg)

	for _, sub := range subs {
		// ! TODO : Один медленный подписчик не должен тормозить остальных (+)
		select {
		case sub.ch <- msg:
		default:
			// if is full
			log.Warnf("%s: append extra message!", loc)
			sub.exMu.Lock()
			sub.extra = append(sub.extra, msg)
			sub.exMu.Unlock()
		}
	}

	return nil
}

func (s *subPub) Close(ctx context.Context) error {
	loc := GLOC_SP + "Close()"

	log.Warnf("%s: close subs", loc)
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
		log.Warnf("%s: close waiting for all routines!", loc)
		return nil
	case <-ctx.Done():
		log.Warnf("%s: forcibly close by context!", loc)
		return ctx.Err()
	}
}
