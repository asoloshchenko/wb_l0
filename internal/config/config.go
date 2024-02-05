package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// TODO: add env tags
type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	DBConn     `yaml:"db_con"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle-timeout" env-default:"60s"`
}

type DBConn struct {
	Address  string `yaml:"address" env:"DB_ADDR" env-default:"localhost"`
	Port     string `yaml:"port" env:"DB_PORT" env-default:"5432"`
	Name     string `yaml:"db_name" env:"DB_NAME" env-default:"pud_test"`
	Username string `yaml:"username" env:"DB_USERNAME" env-required:"true"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-required:"true"`
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
		//slog.Error("Can't open config file", configPath)
		os.Exit(1)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		slog.Error("Can't parse config file:", err)
	}

	return &cfg
}