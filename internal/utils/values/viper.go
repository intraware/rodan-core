package values

import (
	"fmt"
	"log"
	"reflect"
	"regexp"

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
		oldCfg := GetConfig()
		if !reloadEqual(oldCfg.Server, newCfg.Server) ||
			!reloadEqual(oldCfg.App, newCfg.App) {
			if oldCfg.App.EmailRegex != newCfg.App.EmailRegex {
				if newCfg.App.EmailRegex != "" {
					regex, err := regexp.Compile(newCfg.App.EmailRegex)
					if err != nil {
						log.Println("[CONFIG] Invalid email regex:", err)
						return
					}
					newCfg.App.CompiledEmail = regex
				}
			} else {
				newCfg.App.CompiledEmail = oldCfg.App.CompiledEmail
			}
			SetConfig(&newCfg)
			log.Println("[CONFIG] Live reload applied")
		} else {
			log.Println("[CONFIG] Change detected, but no reloadable fields updated")
		}
	})
	return nil
}
