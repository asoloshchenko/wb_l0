package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	DBConn     `yaml:"db_con"`
	Nats       `yaml:"nats"`
}

type Nats struct {
	ClusterID   string `yaml:"cluster_id" env-default:"test-cluster"`
	ClientId    string `yaml:"client_id" env-default:"1"`
	ChanName    string `yaml:"chan_name" env-default:"foo"`
	DurableName string `yaml:"durable_name" env-default:"cache-service"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle-timeout" env-default:"60s"`
}

type DBConn struct {
	DbAddr     string `yaml:"address" env:"DB_ADDR" env-default:"localhost"`
	DbPort     string `yaml:"port" env:"DB_PORT" env-default:"5432"`
	DbName     string `yaml:"db_name" env:"DB_NAME" env-default:"wb"`
	DbUsername string `yaml:"username" env:"DB_USERNAME" env-required:"true"`
	DbPassword string `yaml:"password" env:"DB_PASSWORD" env-required:"true"`
}

// Reads config from YAML file in CONFIG_PATH
// if any errors -> exit(1)
func ReadConfig() *Config {

	err := godotenv.Load()
	if err != nil {
		slog.Error(err.Error())
	}

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		slog.Error("Env variable CONFIG_PATH is not set")
		os.Exit(1)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		slog.Error("Can't open config file", slog.Any("configPath", configPath))
		os.Exit(1)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		slog.Error("Can't parse config file:", slog.Any("configPath", configPath))
		os.Exit(1)
	}

	return &cfg
}
