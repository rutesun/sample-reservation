package config

import (
	"github.com/kelseyhightower/envconfig"
)

type config {
	Port        int `default:"8080"`
	Host 				string `default:"0.0.0.0"`
	Redis struct {
		Host string `default:"0.0.0.0"`
		Port int `default:"6379"`
	}
}

func Parse() (*config, error) {
	c := config{}
	if err := envconfig.Process("", &c); err != nil {
		return nil, err
	}
	return &c, nil
}
