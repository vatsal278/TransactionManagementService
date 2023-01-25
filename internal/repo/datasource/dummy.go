package datasource

import (
	"github.com/PereRohit/util/log"

	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/model"
)

type dummyDs struct {
	dummySvc *config.DummyInternalSvc
}

func NewDummyDs(dummySvc *config.DummyInternalSvc) DataSource {
	return &dummyDs{
		dummySvc: dummySvc,
	}
}

func (d dummyDs) Ping(req *model.PingDs) (*model.DsResponse, error) {
	// internal logic to access datasource, etc.
	log.Info(d.dummySvc.Data)
	return &model.DsResponse{
		Data: req.Data,
	}, nil
}

func (d dummyDs) HealthCheck() bool {
	// actual check for service condition
	log.Debug(d.dummySvc.Data)
	return true
}
