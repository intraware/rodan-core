package config

import (
	"fmt"
	"regexp"
	"time"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Docker   DockerConfig   `mapstructure:"docker"`
	Database DatabaseConfig `mapstructure:"database"`
	App      AppConfig      `mapstructure:"app"`
}

type ServerConfig struct {
	Host       string         `mapstructure:"host"`
	Port       int            `mapstructure:"port"`
	Production bool           `mapstructure:"production"`
	CORSURL    []string       `mapstructure:"cors-url"`
	Security   SecurityConfig `mapstructure:"security"`
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
	BindingHost      string          `mapstructure:"binding-host"`
	PortsMaxRetry    int             `mapstructure:"port-retry-times"`
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
	TeamContainerLimit int               `mapstructure:"team-container-limit"`
	FlagFormat         string            `mapstructure:"flag-format"`
	CacheDuration      time.Duration     `mapstructure:"frontend-cache-duration"`
	EmailRegex         string            `mapstructure:"email-regex"`
	CompiledEmail      *regexp.Regexp    `mapstructure:"-"`
	Ban                BanConfig         `mapstructure:"ban"`
}

type LeaderboardConfig struct {
	User                bool          `mapstructure:"user"`
	Team                bool          `mapstructure:"team"`
	DebounceTimer       time.Duration `mapstructure:"debounce-timer"`
	FullPointsThreshold int           `mapstructure:"full-points-threshold"`
	DecaySharpness      float64       `mapstructure:"decay-sharpness"`
	UserBlackList       []int         `mapstructure:"user-blacklist"`
	TeamBlackList       []int         `mapstructure:"team-blacklist"`
}

type BanConfig struct {
	UserBan            bool          `mapstructure:"enable-user-ban"`
	TeamBan            bool          `mapstructure:"enable-team-ban"`
	InitialBanDuration time.Duration `mapstructure:"initial-ban-duration"`
	BanGrowthFactor    float64       `mapstructure:"ban-growth-factor"`
	MaxBanDuration     time.Duration `mapstructure:"max-ban-duration"`
	TrackHistory       bool          `mapstructure:"track-history"`
}

func (cfg *Config) Validate() error {
	lb := cfg.App.Leaderboard
	if lb.DecaySharpness <= 0 {
		return fmt.Errorf("decay-sharpness must be positive")
	}
	if lb.FullPointsThreshold < 0 {
		return fmt.Errorf("full-points-threshold must be >= 0")
	}
	return nil
}
