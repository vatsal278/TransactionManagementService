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
	cfg := svcCfg.Config{}
    	err := config.LoadFromJson("./configs/config.json", &cfg)
    	if err != nil {
    		log.Error(err)
    		os.Exit(1)
    	}

    	svcInitCfg := svcCfg.InitSvcConfig(cfg)

    	r := router.Register(svcInitCfg)

    	server.Run(r, svcInitCfg.SvrCfg)
}
