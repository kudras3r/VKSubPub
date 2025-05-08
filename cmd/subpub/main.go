package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kudras3r/VKSubPub/internal/grpc"
	"github.com/kudras3r/VKSubPub/internal/service"

	"github.com/kudras3r/VKSubPub/pkg/config"
	"github.com/kudras3r/VKSubPub/pkg/logger"
)

func main() {
	// init config
	cfg := config.MustLoad()

	// init logger
	log := logger.New(cfg.LogLevel)
	log.Info(cfg.PrettyView())

	// init sp service
	sps := service.NewSPService(log, &cfg.SubPub)

	// init server
	srv := grpc.New(log, &cfg.GRPC, sps)

	// run server
	go func() {
		log.Info("running the server")
		if err := srv.Run(); err != nil {
			log.Panic(err)
		}
	}()

	// stop server
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.GRPC.Timeout)*time.Second,
	)

	defer cancel()
	srv.Stop(ctx)

	log.Info("gracefully stopped :)")
}
