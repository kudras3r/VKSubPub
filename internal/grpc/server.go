package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/kudras3r/VKSubPub/internal/service"
	"github.com/kudras3r/VKSubPub/pkg/config"
	"github.com/kudras3r/VKSubPub/pkg/logger"
	pb "github.com/kudras3r/VKSubPub/proto/vk_sp"
	"google.golang.org/grpc"
)

const (
	GLOC_SRV = "internal/grpc/server.go/" // for logging
)

type Server struct {
	pb.UnimplementedPubSubServer

	cfg     *config.GRPCConf
	grpcSrv *grpc.Server
	log     *logger.Logger

	spService *service.SPService
}

func New(
	log *logger.Logger, cfg *config.GRPCConf, sps *service.SPService,
) *Server {
	grpcSrv := grpc.NewServer()
	srv := &Server{
		cfg:       cfg,
		log:       log,
		grpcSrv:   grpcSrv,
		spService: sps,
	}
	pb.RegisterPubSubServer(grpcSrv, srv)
	return srv
}

func (s *Server) Run() error {
	loc := GLOC_SRV + "Run()"

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port))
	if err != nil {
		s.log.Errorf("%s: cannot start the server", loc)
		return ErrCannotListenTcpOn(loc, s.cfg.Port, err)
	}
	s.log.Infof("%s: grpc server started at :%d", loc, s.cfg.Port)

	if err := s.grpcSrv.Serve(lis); err != nil {
		s.log.Errorf("%s: cannot serve the grpc", loc)
		return ErrCannotServeGRPC(loc, err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) {
	loc := GLOC_SRV + "Stop()"

	s.log.Infof("%s: stopping the server", loc)
	s.grpcSrv.Stop()
	s.log.Infof("%s: server stopped", loc)

	s.spService.Close(ctx)
}
