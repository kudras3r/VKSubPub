package grpc

import (
	"fmt"
	"net"

	// "github.com/kudras3r/VKSubPub/internal/services/subpub"
	"github.com/kudras3r/VKSubPub/internal/subpub"
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
	// spService *subpub.SPService

	sp subpub.SubPub
}

func New(
	// log *logger.Logger, cfg *config.GRPCConf, sps *subpub.SPService,
	log *logger.Logger, cfg *config.GRPCConf, sp subpub.SubPub,
) *Server {
	grpcSrv := grpc.NewServer()
	srv := &Server{
		cfg:     cfg,
		log:     log,
		grpcSrv: grpcSrv,
		sp:      sp,
	}
	pb.RegisterPubSubServer(grpcSrv, srv)
	return srv
}

func (s *Server) Run() error {
	loc := GLOC_SRV + "Run()"

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port))
	if err != nil {
		return ErrCannotListenTcpOn(loc, s.cfg.Port, err)
	}
	s.log.Infof("grpc server started at :%d", s.cfg.Port)

	if err := s.grpcSrv.Serve(lis); err != nil {
		return ErrCannotServeGRPC(loc, err)
	}

	return nil
}

func (s *Server) Stop() {

}
