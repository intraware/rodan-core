package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server   ServerConfig   `toml:"server"`
	Security SecurityConfig `toml:"security"`
	Docker   DockerConfig   `toml:"docker"`
	Database DatabaseConfig `toml:"database"`
	App      AppConfig      `toml:"app"`
}

type ServerConfig struct {
	Host       string   `toml:"host"`
	Port       int      `toml:"port"`
	Production bool     `toml:"production"`
	CORSURL    []string `toml:"cors-url"`
}

type SecurityConfig struct {
	JWTSecret  string `toml:"jwt-secret"`
	FlagSecret string `toml:"flag-secret"`
}

type DockerConfig struct {
	SocketURL string          `toml:"socket-url"`
	PortRange DockerPortRange `toml:"port-range"`
}

type DockerPortRange struct {
	Start int `toml:"start"`
	End   int `toml:"end"`
}

type DatabaseConfig struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	Username     string `toml:"username"`
	Password     string `toml:"password"`
	DatabaseName string `toml:"database-name"`
	SSLMode      string `toml:"ssl-mode"`
}

type AppConfig struct {
	TeamCodeRefreshMinutes int               `toml:"team-code-refresh-minutes"`
	EnableLeaderboard      LeaderboardConfig `toml:"enable-leaderboard"`
	TokenExpiry            int               `toml:"token-expiry"`
	TOTPIssuer             string            `toml:"totp-issuer"`
}

type LeaderboardConfig struct {
	User bool `toml:"user"`
	Team bool `toml:"team"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		fmt.Println("The config file is invalid")
		return Config{}, err
	}
	return cfg, nil
}
