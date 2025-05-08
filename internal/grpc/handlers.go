package grpc

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/kudras3r/VKSubPub/proto/vk_sp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	GLOC_HLS = "internal/grpc/handlers.go/" // for logging
)

func (s *Server) Subscribe(
	r *pb.SubscribeRequest,
	st grpc.ServerStreamingServer[pb.Event],
) error {
	loc := GLOC_HLS + "Subscribe()"

	ctx := st.Context()
	msgsCh := make(chan string, s.cfg.MsgsChSize)
	defer close(msgsCh)

	key := r.GetKey()
	s.log.Infof("%s: trying to subscribe at %s", loc, key)

	sub, err := s.spService.Subscribe(ctx, key, msgsCh)
	if err != nil {
		s.log.Errorf("%s: failed to subscribe", loc)
		return status.Error(codes.Internal, err.Error())
	}
	defer func() {
		sub.Unsubscribe()
		s.log.Infof("%s: trying to unsubscribe from key=%s", loc, key)
	}()

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("%s: client unsubscribed from key=%s", loc, key)
			return status.Error(codes.Internal, ctx.Err().Error())

		case m := <-msgsCh:
			if err := st.Send(&pb.Event{Data: m}); err != nil {
				s.log.Errorf("%s: failed to send message: %v", loc, err)
				s.log.Infof("%s: client unsubscribed from key=%s", loc, key)
				return status.Error(codes.Internal, SErrFailedToSendMsg(m))
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

	err := s.spService.Publish(data, key)
	if err != nil {
		s.log.Errorf("%s: failed to publish", loc)
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}

	return &empty.Empty{}, nil
}
