package logic

import (
	"bytes"
	"encoding/json"
	"github.com/PereRohit/util/log"
	respModel "github.com/PereRohit/util/model"
	"github.com/google/uuid"
	"github.com/vatsal278/TransactionManagementService/internal/codes"
	"github.com/vatsal278/TransactionManagementService/internal/config"
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
	NewTransaction(transaction model.Transaction) *respModel.Response
}

type transactionManagementServiceLogic struct {
	DsSvc  datasource.DataSourceI
	AccSvc config.AccSvc
}

func NewTransactionManagementServiceLogic(ds datasource.DataSourceI, as config.AccSvc) TransactionManagementServiceLogicIer {
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
	//var getTransaction model.GetTransaction
	var getTransactions []model.GetTransaction
	if err != nil {
		log.Error(err)
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrGetTransaction),
			Data:    nil,
		}
	}
	for _, s := range transactions {
		getTransaction := model.GetTransaction{
			AccountNumber: s.AccountNumber,
			TransactionId: s.TransactionId,
			Amount:        s.Amount,
			TransferTo:    s.TransferTo,
			CreatedAt:     s.CreatedAt,
			UpdatedAt:     s.UpdatedAt,
			Status:        s.Status,
			Type:          s.Type,
			Comment:       s.Comment,
		}
		getTransactions = append(getTransactions, getTransaction)
	}
	totalPages := int(math.Ceil(float64(count) / float64(limit)))
	nextPage := -1
	if count-offset > limit {
		nextPage = page + 1
	}
	resp := model.PaginatedResponse{Response: getTransactions, Pagination: model.Paginate{CurrentPage: page, NextPage: nextPage, TotalPage: totalPages}}
	return &respModel.Response{
		Status:  http.StatusOK,
		Message: "SUCCESS",
		Data:    resp,
	}
}

func (l transactionManagementServiceLogic) NewTransaction(transaction model.Transaction) *respModel.Response {
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
		req, err := http.NewRequest("PUT", l.AccSvc.Host+":"+l.AccSvc.Port+"/microbank/v1/account"+l.AccSvc.Route, bytes.NewReader(reqBody))
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
