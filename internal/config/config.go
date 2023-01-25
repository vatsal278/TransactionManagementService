package config

import (
	"github.com/PereRohit/util/config"
)

type Config struct {
	ServiceRouteVersion string              `json:"service_route_version"`
	ServerConfig        config.ServerConfig `json:"server_config"`
	// add custom config structs below for any internal services
	DummyCfg DummySvcCfg `json:"custom_svc"`
}

type DummySvcCfg struct {
	DummyCfg string `json:"data"`
}

type SvcConfig struct {
	cfg                 *Config
	ServiceRouteVersion string
	SvrCfg              config.ServerConfig
	// add internal services after init
	DummySvc DummyInternalSvc
}

type DummyInternalSvc struct {
	Data string
}

func InitSvcConfig(cfg Config) *SvcConfig {
	// init required services and assign to the service struct fields
	dummySvc := DummyInternalSvc{
		Data: cfg.DummyCfg.DummyCfg,
	}
	return &SvcConfig{
		cfg:                 &cfg,
		ServiceRouteVersion: cfg.ServiceRouteVersion,
		SvrCfg:              cfg.ServerConfig,
		DummySvc:            dummySvc,
	}
}
