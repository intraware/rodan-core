package config

import (
	"fmt"
	"regexp"
	"time"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Security SecurityConfig `mapstructure:"security"`
	Docker   DockerConfig   `mapstructure:"docker"`
	Database DatabaseConfig `mapstructure:"database"`
	App      AppConfig      `mapstructure:"app"`
}

type ServerConfig struct {
	Host       string   `mapstructure:"host"`
	Port       int      `mapstructure:"port"`
	Production bool     `mapstructure:"production"`
	CORSURL    []string `mapstructure:"cors-url"`
}

type SecurityConfig struct {
	JWTSecret  string `mapstructure:"jwt-secret"`
	FlagSecret string `mapstructure:"flag-secret"`
}

type DockerConfig struct {
	SocketURL        string          `mapstructure:"socket-url"`
	PortRange        DockerPortRange `mapstructure:"port-range"`
	ContainerTimeout time.Duration   `mapstructure:"container-timeout"`
	PoolSize         int             `mapstructure:"pool-size"`
	CleanOrphaned    bool            `mapstructure:"clean-orphaned"`
}

type DockerPortRange struct {
	Start int `mapstructure:"start"`
	End   int `mapstructure:"end"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	DatabaseName string `mapstructure:"database-name"`
	SSLMode      string `mapstructure:"ssl-mode"`
}

type AppConfig struct {
	Leaderboard        LeaderboardConfig `mapstructure:"leaderboard"`
	TokenExpiry        time.Duration     `mapstructure:"token-expiry"`
	TOTPIssuer         string            `mapstructure:"totp-issuer"`
	TeamSize           int               `mapstructure:"team-size"`
	BanMode            string            `mapstructure:"ban-mode"`
	TeamContainerLimit int               `mapstructure:"team-container-limit"`
	FlagFormat         string            `mapstructure:"flag-format"`
	CacheDuration      time.Duration     `mapstructure:"frontend-cache-duration"`
	EmailRegex         string            `mapstructure:"email-regex"`
	CompiledEmail      *regexp.Regexp    `mapstructure:"-"`
}

type LeaderboardConfig struct {
	User                bool          `mapstructure:"user"`
	Team                bool          `mapstructure:"team"`
	DebounceTimer       time.Duration `mapstructure:"debounce-timer"`
	FullPointsThreshold int           `mapstructure:"full-points-threshold"`
	DecaySharpness      float64       `mapstructure:"decay-sharpness"`
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
