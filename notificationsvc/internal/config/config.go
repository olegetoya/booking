package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env   string      `yaml:"env"`
	Kafka KafkaConfig `yaml:"kafka"`
	SMTP  SMTPConfig  `yaml:"smtp"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"group_id"`
}

type SMTPConfig struct {
	Host      string `yaml:"host" env:"SMTP_HOST"`
	Port      string `yaml:"port" env:"SMTP_PORT"`
	Recipient string `yaml:"recipient" env:"SMTP_RECIPIENT"`
}

func MustLoad() *Config {
	var path string

	flag.StringVar(&path, "config", "", "config file path")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	if path == "" {
		panic("config file path is empty")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	if len(cfg.Kafka.Brokers) == 0 ||
		cfg.Kafka.Topic == "" ||
		cfg.Kafka.GroupID == "" {
		panic("incomplete kafka configuration")
	}

	return &cfg
}
