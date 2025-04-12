package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rocvitc/hackvault/server/api"
	"github.com/rocvitc/hackvault/server/config"
	"github.com/rocvitc/hackvault/server/utils"
	"github.com/rocvitc/hackvault/server/utils/middleware"
)

func main() {
	var cfg config.Config
	if config, err := config.LoadConfig("./config.toml"); err != nil {
		fmt.Printf("Error in loading config file: %v\n", err)
		return
	} else {
		cfg = config
	}
	utils.NewLogger(cfg.Prod)
	if cfg.Prod {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())
	api.LoadRoutes(r)
	fmt.Printf("[ENGINE] Server started at %s:%d\n", cfg.Host, cfg.Port)
	r.Run(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
}
