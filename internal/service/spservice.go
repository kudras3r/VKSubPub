package service

import (
	"context"

	"github.com/kudras3r/VKSubPub/internal/subpub"
	"github.com/kudras3r/VKSubPub/pkg/config"
	"github.com/kudras3r/VKSubPub/pkg/logger"
)

const (
	GLOC_SPSER = "internal/services/subpub/spservice.go/" // for logging
)

type SPService struct {
	sp  subpub.SubPub
	log *logger.Logger
}

func NewSPService(
	log *logger.Logger, cfg *config.SPConf,
) *SPService {
	sp := subpub.NewSubPub(log, cfg)
	return &SPService{
		sp:  sp,
		log: log,
	}
}

func keyIsValid(key string) bool {
	// ? THINK : blogic validation
	_ = key
	return true
}

func (s SPService) Subscribe(
	ctx context.Context, key string, msgsCh chan string,
) (subpub.Subscription, error) {
	loc := GLOC_SPSER + "Subscribe()"

	if !keyIsValid(key) {
		s.log.Errorf("%s: invalid key=%s", loc, key)
		return nil, ErrInvalidKey(key)
	}

	sub, err := s.sp.Subscribe(key, func(msg interface{}) {
		if _, ok := msg.(string); ok {
			select {
			case msgsCh <- msg.(string):
			case <-ctx.Done():
				// client leave
			}
		} else {
			s.log.Errorf("%s: received non-string message: %T", loc, msg)
		}
	})
	if err != nil {
		s.log.Errorf("%s: failed to subscribe on key=%s: %v", loc, key, err)
		return nil, ErrFailedToSubscribe(key)
	}

	s.log.Infof("%s: subscribed at %s", loc, key)

	return sub, nil
}

func (s *SPService) Publish(data, key string) error {
	loc := GLOC_SPSER + "Publish()"

	if !keyIsValid(key) {
		s.log.Errorf("%s: invalid key=%s", loc, key)
		return ErrInvalidKey(key)
	}

	err := s.sp.Publish(key, data)
	if err != nil {
		s.log.Errorf("%s: failed to publish on key=%s: %v", loc, key, err)
		return err
	}

	s.log.Infof("%s: published { data: %s, key: %s }", loc, data, key)
	return nil
}

func (s *SPService) Close(ctx context.Context) error {
	loc := GLOC_SPSER + "Close()"

	s.log.Debugf("%s: closing sub-pub service", loc)
	if err := s.sp.Close(ctx); err != nil {
		s.log.Errorf("%s: failed to close sub-pub service: %v", loc, err)
		return err
	}

	s.log.Infof("%s: closed sub-pub service", loc)
	return nil
}
