package handler

import (
	"github.com/PereRohit/util/log"
	"github.com/PereRohit/util/request"
	"github.com/vatsal278/TransactionManagementService/internal/codes"
	"github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/pkg/session"
	"net/http"

	"github.com/PereRohit/util/response"

	"github.com/vatsal278/TransactionManagementService/internal/logic"
	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
)

const TransactionManagementServiceName = "transactionManagementService"

//go:generate mockgen --build_flags=--mod=mod --destination=./../../pkg/mock/mock_handler.go --package=mock github.com/vatsal278/TransactionManagementService/internal/handler TransactionManagementServiceHandler

type TransactionManagementServiceHandler interface {
	HealthChecker
	GetTransactions(w http.ResponseWriter, r *http.Request)
	NewTransaction(w http.ResponseWriter, r *http.Request)
}

type transactionManagementService struct {
	logic logic.TransactionManagementServiceLogicIer
}

func NewTransactionManagementService(ds datasource.DataSourceI) TransactionManagementServiceHandler {
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

func (svc transactionManagementService) GetTransactions(w http.ResponseWriter, r *http.Request) {
	sessionStruct := session.GetSession(r.Context())
	session, ok := sessionStruct.(model.SessionStruct)
	if !ok {
		response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrAssertUserid), nil)
		return
	}
	resp := svc.logic.GetTransactions(session.UserId)
	response.ToJson(w, resp.Status, resp.Message, resp.Data)
}

func (svc transactionManagementService) NewTransaction(w http.ResponseWriter, r *http.Request) {
	sessionStruct := session.GetSession(r.Context())
	session, ok := sessionStruct.(model.SessionStruct)
	if !ok {
		response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrAssertUserid), nil)
		return
	}
	var newTransaction model.NewTransaction
	status, err := request.FromJson(r, &newTransaction)
	if err != nil {
		log.Error(err)
		response.ToJson(w, status, err.Error(), nil)
		return
	}
	newTransaction.UserId = session.UserId
	resp := svc.logic.NewTransaction(newTransaction)
	response.ToJson(w, resp.Status, resp.Message, resp.Data)
}
