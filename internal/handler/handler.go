package handler

import (
	"fmt"
	"net/http"

	"github.com/PereRohit/util/request"
	"github.com/PereRohit/util/response"

	"github.com/vatsal278/TransactionManagementService/internal/logic"
	"github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
)

const TransactionManagementServiceName = "transactionManagementService"

//go:generate mockgen --build_flags=--mod=mod --destination=./../../pkg/mock/mock_handler.go --package=mock github.com/vatsal278/TransactionManagementService/internal/handler TransactionManagementServiceHandler

type TransactionManagementServiceHandler interface {
	HealthChecker
	Ping(w http.ResponseWriter, r *http.Request)
}

type transactionManagementService struct {
	logic logic.TransactionManagementServiceLogicIer
}

func NewTransactionManagementService(ds datasource.DataSource) TransactionManagementServiceHandler {
	svc := &transactionManagementService{
		logic: logic.NewTransactionManagementServiceLogic(ds),
	}
    AddHealthChecker(svc)
	return svc
}

func (svc transactionManagementService) HealthCheck() (svcName string, msg string, stat bool) {
	set := false
	defer func() {
		svcName = TransactionManagementServiceName
		if !set {
			msg = ""
			stat = true
		}
	}()
	stat = svc.logic.HealthCheck()
    set = true
	return
}

func (svc transactionManagementService) Ping(w http.ResponseWriter, r *http.Request) {
	req := &model.PingRequest{}

	suggestedCode, err := request.FromJson(r, req)
	if err != nil {
		response.ToJson(w, suggestedCode, fmt.Sprintf("FAILED: %s", err.Error()), nil)
		return
	}
	// call logic
	resp := svc.logic.Ping(req)
	response.ToJson(w, resp.Status, resp.Message, resp.Data)
	return
}