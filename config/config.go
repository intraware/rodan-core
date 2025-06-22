package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Host                    string `toml:"host"`
	Port                    int16  `toml:"port"`
	Prod                    bool   `toml:"production"`
	JwtSecret               string `toml:"jwt-secret"`
	FlagSecret              string `toml:"flag-secret"`
	CORS                    string `toml:"cors-url"`
	TeamCodeRefreshMinutes  int    `toml:"team-code-refresh-minutes"`
	ContainerPortRangeStart int    `toml:"container-port-range-start"`
	ContainerPortRangeEnd   int    `toml:"container-port-range-end"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		fmt.Println("The config file is invalid")
		return Config{}, err
	}
	return cfg, nil
}
