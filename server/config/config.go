package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Host      string `toml:"host"`
	Port      int16  `toml:"port"`
	Prod      bool   `toml:"production"`
	JwtSecret string `toml:"jwt-secret"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		fmt.Println("The config file is invalid")
		return Config{}, err
	}
	return cfg, nil
}
