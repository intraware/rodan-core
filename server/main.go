package main

import (
	"fmt"

	"github.com/rocvitc/hackvault/server/config"
)

func main() {
	var cfg config.Config
	if config, err := config.LoadConfig("./config.toml"); err != nil {
		fmt.Printf("Error in loading config file: %v\n", err)
		return
	} else {
		cfg = config
	}
	fmt.Println(cfg)
}
