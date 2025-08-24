package models

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/intraware/rodan/internal/config"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {
	dbURL := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.DatabaseName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)
	if envDBURL := os.Getenv("DATABASE_URL"); envDBURL != "" {
		logrus.Warn("DATABASE_URL is set; overriding config values from TOML")
		dbURL = envDBURL
	} else {
		logrus.Info("Using database config from TOML file")
	}
	logLevel := logger.Info
	if cfg.Server.Production {
		logLevel = logger.Silent
	}
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{
			TranslateError: true,
			Logger:         logger.Default.LogMode(logLevel),
		})
		if err == nil {
			break
		}
		logrus.Errorf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Println("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}
	if err != nil {
		logrus.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
	}
	if err := DB.AutoMigrate(&User{}, &Team{}, &Challenge{}, &Container{}, &Solve{}, &BanHistory{}); err != nil {
		logrus.Fatalf("Failed to migrate database: %v", err)
	}
	logrus.Println("Database initialized successfully")
}
