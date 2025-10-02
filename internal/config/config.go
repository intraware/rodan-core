package config

import (
	"fmt"
	"time"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server" reload:"true"`
	Docker   DockerConfig   `mapstructure:"docker"`
	Database DatabaseConfig `mapstructure:"database"`
	App      AppConfig      `mapstructure:"app" reload:"true"`
}

type ServerConfig struct {
	Host       string         `mapstructure:"host"`
	Port       int            `mapstructure:"port"`
	Production bool           `mapstructure:"production" reload:"true"`
	CORSURL    []string       `mapstructure:"cors-url" reload:"true"`
	Security   SecurityConfig `mapstructure:"security" reload:"true"`
}

type SecurityConfig struct {
	JWTSecret  string `mapstructure:"jwt-secret" reload:"true"`
	FlagSecret string `mapstructure:"flag-secret" reload:"true"`
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
	MaxTries     int    `mapstructure:"max-tries"`
}

type AppConfig struct {
	TokenExpiry   time.Duration      `mapstructure:"token-expiry" reload:"true"`
	Leaderboard   LeaderboardConfig  `mapstructure:"leaderboard" reload:"true"`
	FlagFormat    string             `mapstructure:"flag-format" reload:"true"`
	CacheDuration time.Duration      `mapstructure:"frontend-cache-duration" reload:"true"`
	Ban           BanConfig          `mapstructure:"ban" reload:"true"`
	AppCache      CacheConfig        `mapstructure:"cache"`
	Notification  NotificationConfig `mapstructure:"notifications" reload:"true"`
}

type CacheConfig struct {
	InApp                 bool          `mapstructure:"in-app"`
	ServiceUrl            string        `mapstructure:"service-url"`
	ServiceType           string        `mapstructure:"service-type"`
	InternalCacheSize     int           `mapstructure:"internal-cache-size"`
	InternalCacheDuration time.Duration `mapstructure:"internal-cache-duration"`
	SkipLocalCache        bool          `mapstructure:"skip-internal-cache"`
}

type LeaderboardConfig struct {
	User                bool          `mapstructure:"user" reload:"true"`
	Team                bool          `mapstructure:"team" reload:"true"`
	DebounceTimer       time.Duration `mapstructure:"debounce-timer" reload:"true"`
	FullPointsThreshold int           `mapstructure:"full-points-threshold" reload:"true"`
	DecaySharpness      float64       `mapstructure:"decay-sharpness" reload:"true"`
	UserBlackList       []uint        `mapstructure:"user-blacklist" reload:"true"`
	TeamBlackList       []uint        `mapstructure:"team-blacklist" reload:"true"`
}

type BanConfig struct {
	UserBan            bool          `mapstructure:"enable-user-ban" reload:"true"`
	TeamBan            bool          `mapstructure:"enable-team-ban" reload:"true"`
	InitialBanDuration time.Duration `mapstructure:"initial-ban-duration" reload:"true"`
	BanGrowthFactor    float64       `mapstructure:"ban-growth-factor" reload:"true"`
	MaxBanDuration     time.Duration `mapstructure:"max-ban-duration" reload:"true"`
}

type NotificationConfig struct {
	Enabled        bool                     `mapstructure:"enabled" reload:"true"`
	DeliveryMethod string                   `mapstructure:"delivery-method" reload:"true"`
	DefaultRetry   int                      `mapstructure:"default-retry" reload:"true"`
	RetryDelay     time.Duration            `mapstructure:"retry-delay" reload:"true"`
	Timeout        time.Duration            `mapstructure:"timeout" reload:"true"`
	HTTP           *HTTPNotificationConfig  `mapstructure:"http" reload:"true"`
	Kafka          *KafkaNotificationConfig `mapstructure:"kafka" reload:"true"`
}

type HTTPNotificationConfig struct {
	URL          string `mapstructure:"url" reload:"true"`
	Endpoint     string `mapstructure:"endpoint" reload:"true"`
	APIKey       string `mapstructure:"api-key" reload:"true"`
	HashedAPIKey string `mapstructure"-"`
}

type KafkaNotificationConfig struct {
	Brokers []string `mapstructure:"brokers" reload:"true"`
	Topic   string   `mapstructure:"topic" reload:"true"`
	GroupID string   `mapstructure:"group_id" reload:"true"`
}

func (cfg *Config) Validate() error {
	lb := cfg.App.Leaderboard
	if lb.DecaySharpness <= 0 {
		return fmt.Errorf("decay-sharpness must be positive")
	}
	if lb.FullPointsThreshold < 0 {
		return fmt.Errorf("full-points-threshold must be >= 0")
	}
	cache := cfg.App.AppCache
	if cache.InApp {
		if cache.ServiceType != "redis" {
			return fmt.Errorf("only supported service is redis")
		} else if cache.ServiceType == "" {
			return fmt.Errorf("service-url cannot be empty")
		}
	}
	return nil
}
