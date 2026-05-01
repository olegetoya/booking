package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env     string        `yaml:"env"`
	HTTP    HTTPConfig    `yaml:"http"`
	Clients ClientsConfig `yaml:"clients"`
}

type HTTPConfig struct {
	Host              string        `yaml:"host"`
	Port              string        `yaml:"port"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
}

type ClientsConfig struct {
	HotelSvc HotelSvcClientConfig `yaml:"hotelsvc"`
}

type HotelSvcClientConfig struct {
	GRPC GRPCClientConfig `yaml:"grpc"`
}

type GRPCClientConfig struct {
	Host    string        `yaml:"host"`
	Port    string        `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
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
	if res != "" {
		return res
	}

	res = os.Getenv("CONFIG_PATH")

	if res == "" {
		panic("config file path is empty")
	}
	return res
}
