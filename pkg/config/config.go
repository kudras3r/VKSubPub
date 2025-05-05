package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/codingconcepts/env"
	"github.com/joho/godotenv"
)

type Config struct {
	SubPub   SPConf   `env:"SP"`
	GRPC     GRPCConf `env:"GRPC"`
	LogLevel string   `env:"LOG_LEVEL"`
}

type SPConf struct {
	SubQSize       int `env:"SP_SUBQ_SIZE"`
	MaxExSize      int `env:"SP_MAX_EXTRA_SIZE"`
	DefaultSubsCap int `env:"SP_DEF_SUBS_CAP"`
	DefaultExCap   int `env:"SP_DEF_EXTRA_CAP"`
}

type GRPCConf struct {
	Port    int `env:"GRPC_PORT"`
	Timeout int `env:"GRPC_TIMEOUT"`
}

func (c *Config) PrettyView() string {
	return fmt.Sprintf("\n"+`Config {
    LogLevel: %s
    GRPC: {
        Port: %d
        Timeout: %d
    }
    SubPub: {
        SubQSize: %d
        MaxExSize: %d
        DefaultSubsCap: %d
        DefaultExCap: %d
    }
}`,
		c.LogLevel,
		c.GRPC.Port,
		c.GRPC.Timeout,
		c.SubPub.SubQSize,
		c.SubPub.MaxExSize,
		c.SubPub.DefaultSubsCap,
		c.SubPub.DefaultExCap,
	)
}

func fetchPath() string {
	var path string

	flag.StringVar(&path, "config", "", "path to config")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG")
	}

	return path
}

func MustLoad() *Config {
	envPath := fetchPath()
	if err := godotenv.Load(envPath); err != nil {
		panic("error when load config file: " + err.Error())
	}

	var config Config
	var sp SPConf
	var grpc GRPCConf

	if err := env.Set(&sp); err != nil {
		panic("cannot get SubPubConf env vars: " + err.Error())
	}
	if err := env.Set(&grpc); err != nil {
		panic("cannot get grpcConf env vars: " + err.Error())
	}
	config.LogLevel = os.Getenv("LOG_LEVEL")

	config.SubPub = sp
	config.GRPC = grpc

	return &config
}
