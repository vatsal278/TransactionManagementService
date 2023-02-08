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
	"io/ioutil"
	"math"
	"net/http"
	"time"
)

//go:generate mockgen --build_flags=--mod=mod --destination=./../../pkg/mock/mock_logic.go --package=mock github.com/vatsal278/TransactionManagementService/internal/logic TransactionManagementServiceLogicIer

type TransactionManagementServiceLogicIer interface {
	HealthCheck() bool
	GetTransactions(id string, limit int, page int) *respModel.Response
	DownloadTransaction(id string, cookie string) *respModel.Response
	NewTransaction(transaction model.NewTransaction) *respModel.Response
}

type transactionManagementServiceLogic struct {
	DsSvc   datasource.DataSourceI
	UtilSvc config.ExternalSvc
}

func NewTransactionManagementServiceLogic(ds datasource.DataSourceI, ut config.ExternalSvc) TransactionManagementServiceLogicIer {
	return &transactionManagementServiceLogic{
		DsSvc:   ds,
		UtilSvc: ut,
	}
}

func (l transactionManagementServiceLogic) HealthCheck() bool {
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
		req, err := http.NewRequest("PUT", l.UtilSvc.AccSvcUrl+"/microbank/v1/account/update/transaction", bytes.NewReader(reqBody))
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

func (l transactionManagementServiceLogic) DownloadTransaction(id string, cookie string) *respModel.Response {
	transactions, _, err := l.DsSvc.Get(map[string]interface{}{"transaction_id": id}, 0, 0)
	if err != nil {
		log.Error(err)
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrGetTransaction),
			Data:    nil,
		}
	}
	if len(transactions) == 0 {
		log.Error("no transaction with specified transaction_id found")
		return &respModel.Response{
			Status:  http.StatusBadRequest,
			Message: codes.GetErr(codes.ErrGetTransaction),
			Data:    nil,
		}
	}
	req, err := http.NewRequest("GET", l.UtilSvc.UserSvc+"/microbank/v1/user", nil)
	if err != nil {
		log.Error(err)
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrFetchinDataUserSvc),
			Data:    nil,
		}
	}
	req.AddCookie(&http.Cookie{Name: "token", Value: cookie})
	client := http.Client{Timeout: 2 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrFetchinDataUserSvc),
			Data:    nil,
		}
	}
	if response.StatusCode != http.StatusOK {
		log.Info("Status Not OK", response.StatusCode)
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrFetchinDataUserSvc),
			Data:    nil,
		}
	}
	var userResp respModel.Response
	by, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrReadingReqBody),
			Data:    nil,
		}
	}
	err = json.Unmarshal(by, &userResp)
	if err != nil {
		log.Error(err)
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrUnmarshall),
			Data:    nil,
		}
	}
	user, ok := userResp.Data.(map[string]interface{})
	if !ok {
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrAssertResp),
			Data:    nil,
		}
	}
	pdfSvc := l.UtilSvc.PdfSvc.PdfService
	pdf, err := pdfSvc.GeneratePdf(map[string]interface{}{
		"Name":                      user["name"],
		"TransferFromAccountNumber": transactions[0].AccountNumber,
		"TransferToAccountNumber":   transactions[0].TransferTo,
		"TransactionId":             transactions[0].TransactionId,
		"Amount":                    transactions[0].Amount,
		"Date":                      transactions[0].CreatedAt,
		"Status":                    transactions[0].Status,
		"Type":                      transactions[0].Type,
		"Comment":                   transactions[0].Comment,
	}, l.UtilSvc.PdfSvc.UuId)
	if err != nil {
		log.Error(err)
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrPdf),
			Data:    nil,
		}
	}
	return &respModel.Response{
		Status:  http.StatusOK,
		Message: "SUCCESS",
		Data:    pdf,
	}
}
