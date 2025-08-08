package values

import (
	"fmt"
	"log"
	"regexp"

	"github.com/fsnotify/fsnotify"
	"github.com/intraware/rodan/internal/config"
	"github.com/spf13/viper"
)

func InitWithViper(path string) error {
	viper.SetConfigFile(path)
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if cfg.App.EmailRegex != "" {
		regex, err := regexp.Compile(cfg.App.EmailRegex)
		if err != nil {
			return fmt.Errorf("invalid email regex in config: %w", err)
		}
		cfg.App.CompiledEmail = regex
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	SetConfig(&cfg)
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("[CONFIG] Reloading due to change in:", e.Name)
		var newCfg config.Config
		if err := viper.Unmarshal(&newCfg); err != nil {
			log.Println("[CONFIG] Failed to reload config:", err)
			return
		}
		if err := newCfg.Validate(); err != nil {
			log.Println("[CONFIG] Validation failed after reload:", err)
			return
		}
		SetConfig(&newCfg)
		log.Println("[CONFIG] Reloaded config successfully")
	})
	return nil
}
