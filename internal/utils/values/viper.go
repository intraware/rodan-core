package values

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"reflect"

	"github.com/fsnotify/fsnotify"
	"github.com/intraware/rodan/internal/config"
	"github.com/spf13/viper"
)

func reloadEqual(a, b any) bool {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	if va.Kind() == reflect.Ptr {
		if va.IsNil() || vb.IsNil() {
			return va.IsNil() && vb.IsNil()
		}
		va = va.Elem()
		vb = vb.Elem()
	}
	if va.Kind() != reflect.Struct {
		return reflect.DeepEqual(a, b)
	}
	typ := va.Type()
	for i := 0; i < va.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("reload")
		if tag == "true" {
			fa := va.Field(i).Interface()
			fb := vb.Field(i).Interface()
			if !reloadEqual(fa, fb) {
				return false
			}
		}
	}
	return true
}

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
	if cfg.App.Notification.HTTP != nil && cfg.App.Notification.HTTP.APIKey != "" {
		hash := sha256.Sum256([]byte(cfg.App.Notification.HTTP.APIKey))
		cfg.App.Notification.HTTP.HashedAPIKey = hex.EncodeToString(hash[:])
	}
	if cfg.App.Auth.URL != "" && cfg.App.Auth.ApiKey != "" {
		hash := sha256.Sum256([]byte(cfg.App.Auth.ApiKey))
		cfg.App.Auth.HashedAPIKey = hex.EncodeToString(hash[:])
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
		if newCfg.App.Notification.HTTP != nil && newCfg.App.Notification.HTTP.APIKey != "" {
			hash := sha256.Sum256([]byte(newCfg.App.Notification.HTTP.APIKey))
			newCfg.App.Notification.HTTP.HashedAPIKey = hex.EncodeToString(hash[:])
		}
		if newCfg.App.Auth.URL != "" && newCfg.App.Auth.ApiKey != "" {
			hash := sha256.Sum256([]byte(newCfg.App.Auth.ApiKey))
			newCfg.App.Auth.HashedAPIKey = hex.EncodeToString(hash[:])
		}
		if err := newCfg.Validate(); err != nil {
			log.Println("[CONFIG] Validation failed after reload:", err)
			return
		}
		oldCfg := GetConfig()
		if !reloadEqual(oldCfg.Server, newCfg.Server) ||
			!reloadEqual(oldCfg.App, newCfg.App) {
			SetConfig(&newCfg)
			log.Println("[CONFIG] Live reload applied")
		} else {
			log.Println("[CONFIG] Change detected, but no reloadable fields updated")
		}
	})
	return nil
}
