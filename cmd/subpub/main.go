package main

import (
	"fmt"

	"github.com/kudras3r/VKSubPub/internal/subpub"
	"github.com/kudras3r/VKSubPub/pkg/config"
	"github.com/kudras3r/VKSubPub/pkg/logger"
)

func main() {
	// init config
	cfg := config.MustLoad()
	subpub.SetConf(&cfg.SubPub)

	// init logger
	log := logger.New(cfg.LogLevel)
	_ = log

	log.Info(cfg.PrettyView())

	// init app

	// run grpc server

	// ...
	fmt.Println("kek")
}
