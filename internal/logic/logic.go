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
	"math"
	"net/http"
	"time"
)

//go:generate mockgen --build_flags=--mod=mod --destination=./../../pkg/mock/mock_logic.go --package=mock github.com/vatsal278/TransactionManagementService/internal/logic TransactionManagementServiceLogicIer

type TransactionManagementServiceLogicIer interface {
	HealthCheck() bool
	GetTransactions(id string, limit int, page int) *respModel.Response
	NewTransaction(transaction model.NewTransaction) *respModel.Response
}

type transactionManagementServiceLogic struct {
	DsSvc  datasource.DataSourceI
	AccSvc string
}

func NewTransactionManagementServiceLogic(ds datasource.DataSourceI, as string) TransactionManagementServiceLogicIer {
	return &transactionManagementServiceLogic{
		DsSvc:  ds,
		AccSvc: as,
	}
}

func (l transactionManagementServiceLogic) HealthCheck() bool {
	// check all internal services are working fine
	return l.DsSvc.HealthCheck()
}

func (l transactionManagementServiceLogic) GetTransactions(id string, limit int, page int) *respModel.Response {
	offset := (page - 1) * limit
	transactions, count, err := l.DsSvc.Get(map[string]interface{}{"user_id": id}, limit, offset)
	if err != nil {
		log.Error(err)
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrGetTransaction),
			Data:    nil,
		}
	}
	totalPages := int(math.Ceil(float64(count) / float64(limit)))
	nextPage := -1
	if count-offset > limit {
		nextPage = page + 1
	}
	resp := model.PaginatedResponse{Response: transactions, Pagination: model.Paginate{CurrentPage: page, NextPage: nextPage, TotalPage: totalPages}}
	return &respModel.Response{
		Status:  http.StatusOK,
		Message: "SUCCESS",
		Data:    resp,
	}
}

func (l transactionManagementServiceLogic) NewTransaction(newTransaction model.NewTransaction) *respModel.Response {
	transaction := model.Transaction{
		UserId:        newTransaction.UserId,
		AccountNumber: newTransaction.AccountNumber,
		TransactionId: uuid.NewString(),
		Amount:        newTransaction.Amount,
		TransferTo:    newTransaction.TransferTo,
		Status:        newTransaction.Status,
		Type:          newTransaction.Type,
		Comment:       newTransaction.Comment,
	}
	err := l.DsSvc.Insert(transaction)
	if err != nil {
		log.Error(err)
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrNewTransaction),
			Data:    nil,
		}
	}
	if newTransaction.Status != "approved" {
		return &respModel.Response{
			Status:  http.StatusCreated,
			Message: "SUCCESS",
			Data:    nil,
		}
	}
	upTransaction := model.UpdateTransaction{AccountNumber: newTransaction.AccountNumber, Amount: newTransaction.Amount, TransactionType: newTransaction.Type}
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
		req, err := http.NewRequest("PUT", l.AccSvc+"/microbank/v1/account/update/transaction", bytes.NewReader(reqBody))
		if err != nil {
			log.Error(err)
			return
		}
		client := http.Client{Timeout: 3 * time.Second}
		_, err = client.Do(req)
		if err != nil {
			log.Error(err)
			return
		}
	}(by)

	return &respModel.Response{
		Status:  http.StatusCreated,
		Message: "SUCCESS",
		Data:    nil,
	}
}
