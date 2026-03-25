package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env            string        `yaml:"env" env-required:"true"`
	JWTSecret      string        `yaml:"jwt_secret" env-required:"true"`
	ExpirationTime time.Duration `yaml:"exp_time" env-required:"true"`
	StorageURL     string        `yaml:"storage_url" env-required:"true"`
	HTTPServer     HTTPServer    `yaml:"http_server" env-required:"true"`
}

type HTTPServer struct {
	URL         string        `yaml:"url" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"1s"`
}

func MustLoad() *Config {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		log.Fatal("CONFIG_PATH env variable not set")
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		log.Fatalf("file %s does not exist", cfgPath)
	}

	cfg := new(Config)

	if err := cleanenv.ReadConfig(cfgPath, cfg); err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	return cfg
}
