package subpub

import "sync"

type subscription struct {
	subject string
	once    sync.Once
	sub     *subscriber
	sp      *subPub
}

func (s *subscription) Unsubscribe() {
	// ? THINK : O(n) is bad :(
	s.once.Do(func() {
		s.sp.mu.Lock()
		defer s.sp.mu.Unlock()

		subs, ok := s.sp.subcrs[s.subject]
		if !ok {
			return
		}

		for i, sub := range subs {
			if sub == s.sub {
				// remove sub
				s.sp.subcrs[s.subject][i] = subs[len(subs)-1]
				s.sp.subcrs[s.subject] = subs[:len(subs)-1]
				break
			}
		}

		if len(s.sp.subcrs[s.subject]) == 0 {
			delete(s.sp.subcrs, s.subject)
		}

		close(s.sub.stop)
	})
}
