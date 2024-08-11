package config

import (
	"Users/pkg/logging"
	"github.com/ilyakaznacheev/cleanenv"
	"sync"
)

type Config struct {
	Postgres struct {
		Host     string `yaml:"host" env-required:"true"`
		Port     string `yaml:"port" env-required:"true"`
		Database string `yaml:"database" env-required:"true"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"postgres" env-required:"true"`

	GRPC struct {
		IP   string `yaml:"ip"`
		Port int    `yaml:"port"`
	} `yaml:"grpc"`

	HTTP struct {
		IP   string `yaml:"ip"`
		Port int    `yaml:"port"`
		CORS struct {
			AllowedMethods   []string `yaml:"allowed_methods"`
			AllowedOrigins   []string `yaml:"allowed_origins"`
			AllowCredentials bool     `yaml:"allow_credentials"`
			AllowedHeaders   []string `yaml:"allowed_headers"`
			ExposedHeaders   []string `yaml:"exposed_headers"`
		} `yaml:"cors"`
	} `yaml:"http"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("read application config")
		instance = &Config{}
		if err := cleanenv.ReadConfig("config/local.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
	})
	return instance
}
