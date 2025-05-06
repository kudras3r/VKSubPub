package grpc

import (
	"fmt"
	"net"

	"github.com/kudras3r/VKSubPub/pkg/config"
	"github.com/kudras3r/VKSubPub/pkg/logger"
	pb "github.com/kudras3r/VKSubPub/proto/vk_sp"
	"google.golang.org/grpc"
)

const (
	GLOC_SRV = "internal/grpc/server.go/"
)

type Server struct {
	pb.UnimplementedPubSubServer
	cfg     *config.GRPCConf
	grpcSrv *grpc.Server
	log     *logger.Logger
}

func New(
	log *logger.Logger, cfg *config.GRPCConf,
) *Server {
	srv := grpc.NewServer()
	pb.RegisterPubSubServer(srv, &Server{})
	return &Server{
		cfg:     cfg,
		log:     log,
		grpcSrv: srv,
	}
}

func (s *Server) Run() error {
	loc := GLOC_SRV + "MustRun()"

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port))
	if err != nil {
		return ErrCannotListenTcpOn(loc, s.cfg.Port, err)
	}
	if err := s.grpcSrv.Serve(lis); err != nil {
		return ErrCannotServeGRPC(loc, err)
	}

	return nil
}

func (s *Server) Stop() {

}
