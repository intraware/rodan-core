package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api"
	"github.com/intraware/rodan/internal/models"
	"github.com/intraware/rodan/internal/utils"
	"github.com/intraware/rodan/internal/utils/middleware"
	"github.com/intraware/rodan/internal/utils/values"
)

func Run() {
	config_file := os.Getenv("CONFIG_FILE")
	if err := values.InitWithViper(config_file); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	cfg := values.GetConfig()
	models.InitDB(cfg)
	utils.NewLogger(cfg.Server.Production)
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
	fmt.Printf("[ENGINE] Server started at %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	r.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
}
