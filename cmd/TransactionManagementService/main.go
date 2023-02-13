package main

import (
	"os"

	"github.com/PereRohit/util/config"
	"github.com/PereRohit/util/log"
	"github.com/PereRohit/util/server"

	svcCfg "github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/router"
)

func main() {
	// Load the configuration from a JSON file
	cfg := svcCfg.Config{}
	err := config.LoadFromJson("./configs/config.json", &cfg)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// Initialize the service configuration based on the loaded configuration
	svcInitCfg := svcCfg.InitSvcConfig(cfg)

	// Register the routes and handlers for the service
	r := router.Register(svcInitCfg)

	// Start the server with the registered router and server configuration
	server.Run(r, svcInitCfg.SvrCfg)
}
