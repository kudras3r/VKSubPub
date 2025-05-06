package subpub

import (
	"github.com/kudras3r/VKSubPub/internal/subpub"
)

type SPService struct {
	sp subpub.SubPub
}

func (s *SPService) KeyIsValid(key string) bool {
	// ? THINK
	return true
}
