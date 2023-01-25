package logic

import (
	"net/http"

	"github.com/PereRohit/util/log"
    respModel "github.com/PereRohit/util/model"

	"github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
)

//go:generate mockgen --build_flags=--mod=mod --destination=./../../pkg/mock/mock_logic.go --package=mock github.com/vatsal278/TransactionManagementService/internal/logic TransactionManagementServiceLogicIer

type TransactionManagementServiceLogicIer interface {
	Ping(*model.PingRequest) *respModel.Response
    HealthCheck() bool
}

type transactionManagementServiceLogic struct{
	dummyDsSvc datasource.DataSource
}

func NewTransactionManagementServiceLogic(ds datasource.DataSource) TransactionManagementServiceLogicIer {
	return &transactionManagementServiceLogic{
		dummyDsSvc: ds,
    }
}

func (l transactionManagementServiceLogic) Ping(req *model.PingRequest) *respModel.Response {
	// add business logic here
	res, err := l.dummyDsSvc.Ping(&model.PingDs{
    	Data: req.Data,
    })
    if err != nil {
        log.Error("datasource error", err)
    	return &respModel.Response{
    		Status:  http.StatusInternalServerError,
    		Message: "",
    		Data:    nil,
    	}
    }
    return &respModel.Response{
    	Status:  http.StatusOK,
    	Message: "Pong",
    	Data:    res,
    }
}

func (l transactionManagementServiceLogic) HealthCheck() bool {
	// check all internal services are working fine
	return l.dummyDsSvc.HealthCheck()
}