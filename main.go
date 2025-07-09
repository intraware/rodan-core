//go:generate swag init
package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api"
	"github.com/intraware/rodan/config"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/intraware/rodan/utils/middleware"
	"github.com/intraware/rodan/utils/values"
	"github.com/sirupsen/logrus"
)

func main() {
	var cfg config.Config
	if config, err := config.LoadConfig("./sample.config.toml"); err != nil {
		fmt.Printf("Error in loading config file: %v\n", err)
		return
	} else {
		cfg = config
	}
	values.SetConfig(&cfg)
	models.InitDB(&cfg)
	cleanupService, err := utils.NewCleanupService()
	if err != nil {
		logrus.Warnf("Failed to initialize cleanup service: %v", err)
	} else {
		cleanupService.StartCleanupRoutine()
		defer cleanupService.Close()
	}
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
	values.SetRouter(r)
	fmt.Printf("[ENGINE] Server started at %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	r.Run(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
}
