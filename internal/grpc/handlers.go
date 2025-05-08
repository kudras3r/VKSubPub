package grpc

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/kudras3r/VKSubPub/proto/vk_sp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Subscribe(
	r *pb.SubscribeRequest,
	st grpc.ServerStreamingServer[pb.Event],
) error {
	loc := GLOC_SRV + "Subscribe()"

	ctx := st.Context()
	msgCh := make(chan string) // ? THINK buff or no

	key := r.GetKey()
	s.log.Infof("%s: trying to subscribe at %s", loc, key)

	sub, err := s.sp.Subscribe(key, func(msg interface{}) {
		if _, ok := msg.(string); ok {
			select {
			case msgCh <- msg.(string):
			case <-ctx.Done():
				// client leave
			}
		} else {
			s.log.Errorf("%s: received non-string message: %T", loc, msg)
		}
	})
	if err != nil {
		s.log.Errorf("%s: failed to subscribe on key=%s: %v", loc, key, err)
		return status.Error(codes.Internal, SErrFailedToSubscribe)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("%s: client unsubscribed from key=%s", loc, key)
			return ctx.Err()

		case m := <-msgCh:
			if err := st.Send(&pb.Event{Data: m}); err != nil {
				s.log.Errorf("%s: failed to send message: %v", loc, err)
				s.log.Infof("%s: client unsubscribed from key=%s", loc, key)
				return status.Error(codes.Internal, SErrFailedToSendMsg)
			}
		}
	}
}

func (s *Server) Publish(
	ctx context.Context, r *pb.PublishRequest,
) (*empty.Empty, error) {
	loc := GLOC_SRV + "Publish()"

	data, key := r.GetData(), r.GetKey()
	s.log.Infof("%s: trying to publish { data: %s, key: %s }", loc, data, key)

	err := s.sp.Publish(key, data)
	if err != nil {
		s.log.Errorf("%s: failed to publish", loc)
		return &empty.Empty{}, status.Error(codes.Internal, SErrFailedToPublish)
	}

	return &empty.Empty{}, nil
}
