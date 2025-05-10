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

	"github.com/kudras3r/VKSubPub/pkg/config"
	"github.com/kudras3r/VKSubPub/pkg/logger"
)

const (
	GLOC_SP = "internal/subpub/spimpl.go/" // for logging
)

type subscriber struct {
	ch   chan interface{}
	hl   MessageHandler
	stop chan struct{}
	log  *logger.Logger

	exMu  sync.Mutex
	extra []interface{}
}

// For 1 write : M reads (IO bound)
type subPub struct {
	subcrs map[string][]*subscriber
	cfg    *config.SPConf
	log    *logger.Logger

	mu sync.RWMutex
	wg sync.WaitGroup
}

func emptySubject(subject string) bool {
	return subject == ""
}

func nilHandler(hl MessageHandler) bool {
	return hl == nil
}

func (s *subscriber) drainExtra(maxExSize int) {
	loc := GLOC_SP + "drainExtra()"

	s.exMu.Lock()
	defer s.exMu.Unlock()

	if len(s.extra) > maxExSize {
		s.log.Errorf("%s: extra slice size=%d when max=%d!", loc, len(s.extra), maxExSize)
		// ? THINK : what we gonna do
	}

	for len(s.extra) > 0 {
		s.log.Warnf("%s: have extra messages!", loc)
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
		s.log.Warnf("%s: emtry subject!", loc)
		return nil, ErrEmptySubject(loc)
	}
	if nilHandler(cb) {
		s.log.Warnf("%s: handler is nil!", loc)
		return nil, ErrEmptyHandler(loc)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subcrs[subject]; !ok {
		s.subcrs[subject] = make([]*subscriber, 0, s.cfg.DefaultSubsCap)
	}

	nsub := &subscriber{
		ch:   make(chan interface{}, s.cfg.SubQSize),
		hl:   cb,
		stop: make(chan struct{}),
		log:  s.log,

		extra: make([]interface{}, 0, s.cfg.DefaultExCap),
	}
	s.subcrs[subject] = append(s.subcrs[subject], nsub)
	s.log.Infof("%s: new sub added to subject: %s", loc, subject)

	s.wg.Add(1)

	// listen
	go func() {
		defer s.wg.Done()
		for {
			select {
			case msg := <-nsub.ch:
				// ? THINK : may be recover
				nsub.hl(msg) // if hl panic - goroutine will be dead
				nsub.drainExtra(s.cfg.MaxExSize)
			case <-nsub.stop:
				s.log.Debugf("%s: stop sub with subject: %s", loc, subject)
				return
			}
		}
	}()

	return &subscription{
		subject: subject,
		sub:     nsub,
		sp:      s,
		log:     s.log,
	}, nil
}

// Publish publishes the msg argument to the given subject.
func (s *subPub) Publish(subject string, msg interface{}) error {
	loc := GLOC_SP + "Publish()"

	if emptySubject(subject) {
		s.log.Warnf("%s: empty subject!", loc)
		return ErrEmptySubject(loc)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	subs, ok := s.subcrs[subject]
	if !ok {
		s.log.Warnf("%s: no subscribers for subject %s!", loc, subject)
		return nil
	}

	for _, sub := range subs {
		if len(sub.extra) > 0 {
			sub.extra = append(sub.extra, msg)
			continue
		}

		select {
		case sub.ch <- msg:
		default:
			s.log.Warnf("%s: append extra message!", loc)
			sub.exMu.Lock()
			sub.extra = append(sub.extra, msg)
			sub.exMu.Unlock()
		}
	}

	return nil
}

// Close will shutdown sub-pub system.
// May be blocked by data delivery until the context is canceled.
func (s *subPub) Close(ctx context.Context) error {
	loc := GLOC_SP + "Close()"

	done := make(chan struct{})
	go func() {
		s.log.Infof("%s: waiting for all routines to finish", loc)
		s.wg.Wait()
		s.log.Infof("%s: all routines finished", loc)
		close(done)
	}()

	select {
	case <-done:
		s.log.Warnf("%s: close waiting for all routines!", loc)
		return nil
	case <-ctx.Done():
		s.log.Warnf("%s: forcibly close by context!", loc)
		return ctx.Err()
	}
}
