package config

import (
	"fmt"
	"time"
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
	ContainerTimeout time.Duration   `toml:"container-timeout"`
	PoolSize         int             `toml:"pool-size"`
	CleanOrphaned    bool            `toml:"clean-orphaned"`
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
	TokenExpiry        time.Duration     `toml:"token-expiry"`
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

func (cfg *Config) Validate() error {
	lb := cfg.App.Leaderboard
	if lb.DecaySharpness <= 0 {
		return fmt.Errorf("decay-sharpness must be positive")
	}
	if lb.FullPointsThreshold < 0 {
		return fmt.Errorf("full-points-threshold must be >= 0")
	}
	if cfg.App.BanMode != "user" && cfg.App.BanMode != "team" {
		return fmt.Errorf("ban-mode must be 'user' or 'team'")
	}
	return nil
}
