package config

import (
	"testing"

	"github.com/PereRohit/util/config"
	"github.com/PereRohit/util/testutil"
)

func TestInitSvcConfig(t *testing.T) {
	type args struct {
		cfg Config
	}
	tests := []struct {
		name string
		args args
		want *SvcConfig
	}{
		{
			name: "Success",
			args: args{
				cfg: Config{
					ServiceRouteVersion: "v2",
					ServerConfig:        config.ServerConfig{},
					DummyCfg: DummySvcCfg{
						DummyCfg: "dummy cfg",
					},
				},
			},
			want: &SvcConfig{
				cfg: &Config{
					ServiceRouteVersion: "v2",
					ServerConfig:        config.ServerConfig{},
					DummyCfg: DummySvcCfg{
						DummyCfg: "dummy cfg",
					},
				},
				ServiceRouteVersion: "v2",
				SvrCfg:              config.ServerConfig{},
				DummySvc: DummyInternalSvc{
					Data: "dummy cfg",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InitSvcConfig(tt.args.cfg)
			diff := testutil.Diff(got, tt.want)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})
	}
}
