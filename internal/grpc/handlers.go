package grpc

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/kudras3r/VKSubPub/proto/vk_sp"
	"google.golang.org/grpc"
)

func (s *Server) Subscribe(
	r *pb.SubscribeRequest,
	st grpc.ServerStreamingServer[pb.Event],
) error {
	ctx := st.Context()

	msgCh := make(chan string)

	_, err := s.sp.Subscribe(r.Key, func(msg interface{}) {
		if _, ok := msg.(string); !ok {
			return
		}
		select {
		case msgCh <- msg.(string):
		case <-ctx.Done():
		}
	})
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("client unsubscribed from key=%s", r.Key)
			return ctx.Err()
		case m := <-msgCh:
			if err := st.Send(&pb.Event{Data: m}); err != nil {
				s.log.Errorf("failed to send message: %v", err)
				return err
			}
		}
	}
}

func (s *Server) Publish(
	ctx context.Context,
	r *pb.PublishRequest,
) (*empty.Empty, error) {
	err := s.sp.Publish(r.Key, r.Data)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}
