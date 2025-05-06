package grpc

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/kudras3r/VKSubPub/proto/vk_sp"
	"google.golang.org/grpc"
)

func (s *Server) Subscribe(
	*pb.SubscribeRequest,
	grpc.ServerStreamingServer[pb.Event],
) error {
	panic("asd")
}

func (s *Server) Publish(
	context.Context, *pb.PublishRequest,
) (*empty.Empty, error) {
	panic("asd")
}
