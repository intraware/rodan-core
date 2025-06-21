package models

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// Load .env vars
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, using environment variables")
	}

	// Get db URL from env vars
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logrus.Fatalf("DATABASE_URL environment variable not set")
	}

	logrus.Printf("Connecting to database using DATABASE_URL")

	// Connecting to db
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
		if err == nil {
			break
		}
		logrus.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Println("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}

	if err != nil {
		logrus.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
	}

	// Add tables here ...
	if err := DB.AutoMigrate(&User{}, &Team{}, &StaticChallenge{}, &DynamicChallenge{}, &Container{}, &Solve{}); err != nil {
		logrus.Fatalf("Failed to migrate database: %v", err)
	}

	logrus.Println("Database initialized successfully")
}
