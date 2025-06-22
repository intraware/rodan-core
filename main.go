package main

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api"
	"github.com/intraware/rodan/config"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/intraware/rodan/utils/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	var cfg config.Config
	if config, err := config.LoadConfig("./config.toml"); err != nil {
		fmt.Printf("Error in loading config file: %v\n", err)
		return
	} else {
		cfg = config
	}

	// Initialize the database first
	models.InitDB()

	// Initialize and start cleanup service
	cleanupService, err := utils.NewCleanupService()
	if err != nil {
		logrus.Warnf("Failed to initialize cleanup service: %v", err)
	} else {
		cleanupService.StartCleanupRoutine()
		defer cleanupService.Close()
	}

	utils.NewLogger(cfg.Prod)
	if cfg.Prod {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return !cfg.Prod || origin == cfg.CORS
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Accept", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           4 * time.Hour,
	}))

	r.Use(gin.Recovery())
	api.LoadRoutes(r)
	fmt.Printf("[ENGINE] Server started at %s:%d\n", cfg.Host, cfg.Port)
	r.Run(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
}
