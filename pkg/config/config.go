package config

import (
    "log"
    "os"
    "path/filepath"

    "github.com/codingconcepts/env"
    "github.com/joho/godotenv"
)

// ! CHECK
type Config struct {
    SUBCONF1 CONF1  `env:"CONF1"`
    SUBCONF2 CONF2  `env:"CONF2"`
    LogLevel string `env:"LOG_LEVEL"`
}

// ! CHECK
type CONF1 struct {
    F1 string `env:"CONF1_F1"`
}

// ! CHECK
type CONF2 struct {
    F1 string `env:"CONF2_F1"`
}

func Load() *Config {
    envPath := filepath.Join("..", ".", ".env") // ! CHECK
    if err := godotenv.Load(envPath); err != nil {
        log.Fatal(err)
    }

    var config Config
    var SCONF1 CONF1 // ! CHECK
    var SCONF2 CONF2 // ! CHECK

    if err := env.Set(&SCONF1); err != nil {
        log.Fatal("cannot get CONF1 env vars: ", err) // ! CHECK
    }
    if err := env.Set(&SCONF2); err != nil {
        log.Fatal("cannot get CONF2 env vars: ", err) // ! CHECK
    }
    config.LogLevel = os.Getenv("LOG_LEVEL")

    config.SUBCONF1 = SCONF1 // ! CHECK
    config.SUBCONF2 = SCONF2 // ! CHECK

    return &config
}
