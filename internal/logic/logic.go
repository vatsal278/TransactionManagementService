package logic

import (
	"bytes"
	"encoding/json"
	"github.com/PereRohit/util/log"
	respModel "github.com/PereRohit/util/model"
	"github.com/google/uuid"
	"github.com/vatsal278/TransactionManagementService/internal/codes"
	"github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
	"net/http"
	"time"
)

//go:generate mockgen --build_flags=--mod=mod --destination=./../../pkg/mock/mock_logic.go --package=mock github.com/vatsal278/TransactionManagementService/internal/logic TransactionManagementServiceLogicIer

type TransactionManagementServiceLogicIer interface {
	HealthCheck() bool
	GetTransactions(id string) *respModel.Response
	NewTransaction(transaction model.NewTransaction) *respModel.Response
}

type transactionManagementServiceLogic struct {
	DsSvc datasource.DataSourceI
}

func NewTransactionManagementServiceLogic(ds datasource.DataSourceI) TransactionManagementServiceLogicIer {
	return &transactionManagementServiceLogic{
		DsSvc: ds,
	}
}

func (l transactionManagementServiceLogic) HealthCheck() bool {
	// check all internal services are working fine
	return l.DsSvc.HealthCheck()
}

func (l transactionManagementServiceLogic) GetTransactions(id string) *respModel.Response {
	transactions, err := l.DsSvc.Get(map[string]interface{}{"user_id": id}, 10, 0)
	if err != nil {
		log.Error(err)
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrGetTransaction),
			Data:    nil,
		}
	}
	if len(transactions) == 0 {
		return &respModel.Response{
			Status:  http.StatusBadRequest,
			Message: codes.GetErr(codes.ErrNoTransaction),
			Data:    nil,
		}
	}
	resp := model.GetTransaction{}
	return &respModel.Response{
		Status:  http.StatusOK,
		Message: "SUCCESS",
		Data:    resp,
	}
}

func (l transactionManagementServiceLogic) NewTransaction(transaction model.NewTransaction) *respModel.Response {
	transaction.TransactionId = uuid.NewString()
	err := l.DsSvc.Insert(transaction)
	if err != nil {
		log.Error(codes.GetErr(codes.ErrNewTransaction))
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrNewTransaction),
			Data:    nil,
		}
	}
	upTransaction := model.UpdateTransaction{AccountNumber: transaction.AccountNumber, Amount: transaction.Amount, TransactionType: transaction.Type}
	by, err := json.Marshal(upTransaction)
	if err != nil {
		log.Error(codes.GetErr(codes.ErrNewTransaction))
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrNewTransaction),
			Data:    nil,
		}
	}
	go func(reqBody []byte) {
		req, err := http.NewRequest("PUT", "http://localhost:9080/microbank/v1/account/update/transaction", bytes.NewReader(reqBody))
		if err != nil {
			log.Error(err)
			return
		}
		client := http.Client{Timeout: 3 * time.Second}
		client.Do(req)
	}(by)
	return &respModel.Response{
		Status:  http.StatusCreated,
		Message: "SUCCESS",
		Data:    nil,
	}
}
