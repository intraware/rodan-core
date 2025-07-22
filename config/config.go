package config

import (
	"fmt"
	"log"
	"time"

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
	SocketURL        string          `toml:"socket-url"`
	PortRange        DockerPortRange `toml:"port-range"`
	ContainerTimeout int             `toml:"container-timeout"`
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
	Leaderboard        LeaderboardConfig `toml:"leaderboard"`
	TokenExpiry        int               `toml:"token-expiry"`
	TOTPIssuer         string            `toml:"totp-issuer"`
	TeamSize           int               `toml:"team-size"`
	BanMode            string            `toml:"ban-mode"`
	TeamContainerLimit int               `toml:"team-container-limit"`
	FlagFormat         string            `toml:"flag-format"`
}

type LeaderboardConfig struct {
	User                bool          `toml:"user"`
	Team                bool          `toml:"team"`
	DebounceTimer       time.Duration `toml:"debounce-timer"`
	FullPointsThreshold int           `toml:"full-points-threshold"`
	DecaySharpness      float64       `toml:"decay-sharpness"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		fmt.Println("The config file is invalid")
		return Config{}, err
	}
	if cfg.App.BanMode != "user" && cfg.App.BanMode != "team" {
		log.Fatal("Invalid ban-mode. Must be 'user' or 'team'")
	}
	return cfg, nil
}
