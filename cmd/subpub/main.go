package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/kudras3r/VKSubPub/internal/grpc"
	"github.com/kudras3r/VKSubPub/internal/subpub"
	"github.com/kudras3r/VKSubPub/pkg/config"
	"github.com/kudras3r/VKSubPub/pkg/logger"
)

func main() {
	// init config
	cfg := config.MustLoad()

	// init logger
	log := logger.New(cfg.LogLevel)
	log.Info(cfg.PrettyView())

	// init sp
	sp := subpub.NewSubPub()
	subpub.SetConf(&cfg.SubPub)
	subpub.SetLogger(log)

	// init server
	srv := grpc.New(log, &cfg.GRPC, sp)

	// run server
	go func() {
		log.Info("running the server")
		if err := srv.Run(); err != nil {
			log.Panic(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	srv.Stop()
}
