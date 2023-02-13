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

// TransactionManagementServiceLogicIer defines the interface for the transaction management service logic
type TransactionManagementServiceLogicIer interface {
	HealthCheck() bool
	GetTransactions(id string, limit int, page int) *respModel.Response
	DownloadTransaction(id string, cookie string) *respModel.Response
	NewTransaction(transaction model.NewTransaction) *respModel.Response
}

// transactionManagementServiceLogic implements the logic for the transaction management service
type transactionManagementServiceLogic struct {
	DsSvc   datasource.DataSourceI
	UtilSvc config.ExternalSvc
}

// NewTransactionManagementServiceLogic creates a new instance of the transactionManagementServiceLogic
func NewTransactionManagementServiceLogic(ds datasource.DataSourceI, ut config.ExternalSvc) TransactionManagementServiceLogicIer {
	return &transactionManagementServiceLogic{
		DsSvc:   ds,
		UtilSvc: ut,
	}
}

// HealthCheck checks the health of the data source service
func (l transactionManagementServiceLogic) HealthCheck() bool {
	return l.DsSvc.HealthCheck()
}

// GetTransactions retrieves the transactions for the given user id, limit, and page
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

// NewTransaction creates a new transaction and updates the account service if status is "approved"
func (l transactionManagementServiceLogic) NewTransaction(newTransaction model.NewTransaction) *respModel.Response {
	// Create a new transaction using the input data
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

	// Insert the new transaction into the database
	err := l.DsSvc.Insert(transaction)
	if err != nil {
		log.Error(err)
		// If there is an error inserting the transaction, return an error response
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrNewTransaction),
			Data:    nil,
		}
	}

	// If the status of the new transaction is not "approved", return a success response
	if newTransaction.Status != "approved" {
		return &respModel.Response{
			Status:  http.StatusCreated,
			Message: "SUCCESS",
			Data:    nil,
		}
	}

	// If the status of the new transaction is "approved", update the account service asynchronously
	upTransaction := model.UpdateTransaction{AccountNumber: newTransaction.AccountNumber, Amount: newTransaction.Amount, TransactionType: newTransaction.Type}
	by, err := json.Marshal(upTransaction)
	if err != nil {
		log.Error(err)
		// If there is an error marshaling the update transaction, return an error response
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
	// Return a success response
	return &respModel.Response{
		Status:  http.StatusCreated,
		Message: "SUCCESS",
		Data:    nil,
	}
}

// DownloadTransaction is a method of the transactionManagementServiceLogic struct that downloads a transaction as a PDF.
func (l transactionManagementServiceLogic) DownloadTransaction(id string, cookie string) *respModel.Response {
	// Get the transaction with the specified ID from the data store.
	transactions, _, err := l.DsSvc.Get(map[string]interface{}{"transaction_id": id}, 0, 0)
	if err != nil {
		log.Error(err)
		// If an error occurred, return an internal server error response.
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrGetTransaction),
			Data:    nil,
		}
	}
	// If no transaction was found with the specified ID, return a bad request response.
	if len(transactions) == 0 {
		log.Error("no transaction with specified transaction_id found")
		return &respModel.Response{
			Status:  http.StatusBadRequest,
			Message: codes.GetErr(codes.ErrGetTransaction),
			Data:    nil,
		}
	}
	// Create a new HTTP request to the user service to fetch user data.
	req, err := http.NewRequest("GET", l.UtilSvc.UserSvc+"/microbank/v1/user", nil)
	if err != nil {
		log.Error(err)
		// If an error occurred, return an internal server error response.
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrFetchinDataUserSvc),
			Data:    nil,
		}
	}
	// Add the user's authentication token to the request.
	req.AddCookie(&http.Cookie{Name: "token", Value: cookie})
	// Create an HTTP client with a timeout of 3 seconds.
	client := http.Client{Timeout: 3 * time.Second}
	// Send the request to the user service.
	response, err := client.Do(req)
	if err != nil {
		log.Error(err)
		// If an error occurred, return an internal server error response.
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrFetchinDataUserSvc),
			Data:    nil,
		}
	}
	// If the user service did not return an OK status code, return an internal server error response.
	if response.StatusCode != http.StatusOK {
		log.Info("Status Not OK", response.StatusCode)
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrFetchinDataUserSvc),
			Data:    nil,
		}
	}
	// Parse the response body into a response model.
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
	// Assert that the response data is a map.
	user, ok := userResp.Data.(map[string]interface{})
	if !ok {
		return &respModel.Response{
			Status:  http.StatusInternalServerError,
			Message: codes.GetErr(codes.ErrAssertResp),
			Data:    nil,
		}
	}
	// Generate a PDF with the transaction and user data.
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
	// Return a success response
	return &respModel.Response{
		Status:  http.StatusOK,
		Message: "SUCCESS",
		Data:    pdf,
	}
}
