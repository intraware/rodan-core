package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api"
	"github.com/intraware/rodan/internal/cache"
	"github.com/intraware/rodan/internal/models"
	"github.com/intraware/rodan/internal/utils"
	"github.com/intraware/rodan/internal/utils/docker"
	"github.com/intraware/rodan/internal/utils/middleware"
	"github.com/intraware/rodan/internal/utils/values"
)

func Run() {
	configFile := os.Getenv("CONFIG_FILE")
	if err := values.InitWithViper(configFile); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	cfg := values.GetConfig()
	ctx := context.Background()
	models.InitDB(cfg)
	utils.NewLogger(cfg.Server.Production)
	if err := docker.SetupDockerClient(); err != nil {
		log.Fatalf("Failed to setup Docker client: %v", err)
	}
	if cfg.Server.Production {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.Server))
	r.Use(gin.Recovery())
	api.LoadRoutes(r)
	if !cfg.App.AppCache.InApp {
		cache.InitRedis(ctx)
	}
	fmt.Printf("[ENGINE] Server started at %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	r.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
}
