package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env      string     `yaml:"env" env-default:"local"`
	Database DBConfig   `yaml:"database" env-required:"true"`
	GRPC     GRPCConfig `yaml:"grpc" env-required:"true"`
	HTTP     HTTPConfig `yaml:"http" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type DBConfig struct {
	DSN string `yaml:"dsn" env-required:"true"`
}

type HTTPConfig struct {
	Port              int           `yaml:"port" env-required:"true"`
	ReadHeaderTimeout time.Duration `yaml:"readHeaderTimeout" env-default:"10s"`
	ReadTimeout       time.Duration `yaml:"readTimeout" env-default:"10s"`
	WriteTimeout      time.Duration `yaml:"writeTimeout" env-default:"10s"`
	IdleTimeout       time.Duration `yaml:"idleTimeout" env-default:"10s"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config file path is empty")
	}

	// os.Stat проверяет, существует ли файл по такому пути
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist" + path)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "config file path")
	flag.Parse()

	if res == "" {
		panic("config file path is empty")
	}
	return res
}
