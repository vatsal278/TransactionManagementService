package handler

import (
	"fmt"
	"github.com/PereRohit/util/log"
	"github.com/PereRohit/util/request"
	"github.com/PereRohit/util/response"
	"github.com/gorilla/mux"
	"github.com/vatsal278/TransactionManagementService/internal/codes"
	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/pkg/session"
	"net/http"
	"strconv"

	"github.com/vatsal278/TransactionManagementService/internal/logic"
	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
)

const TransactionManagementServiceName = "transactionManagementService"

//go:generate mockgen --build_flags=--mod=mod --destination=./../../pkg/mock/mock_handler.go --package=mock github.com/vatsal278/TransactionManagementService/internal/handler TransactionManagementServiceHandler

// TransactionManagementServiceHandler defines the interface for the Transaction Management Service.
// It defines the methods that can be implemented by the transactionManagementService struct.
type TransactionManagementServiceHandler interface {
	HealthChecker
	GetTransactions(w http.ResponseWriter, r *http.Request)
	NewTransaction(w http.ResponseWriter, r *http.Request)
	DownloadTransaction(w http.ResponseWriter, r *http.Request)
}

// transactionManagementService implements TransactionManagementServiceHandler.
// It has a single field, logic of type logic.TransactionManagementServiceLogicIer, that is used to execute business logic.
type transactionManagementService struct {
	logic logic.TransactionManagementServiceLogicIer
}

// NewTransactionManagementService is a factory method that returns a new TransactionManagementServiceHandler
// It creates a new transactionManagementService and returns it after registering the service with the global health checker.
func NewTransactionManagementService(ds datasource.DataSourceI, ut config.ExternalSvc) TransactionManagementServiceHandler {
	svc := &transactionManagementService{
		logic: logic.NewTransactionManagementServiceLogic(ds, ut),
	}
	AddHealthChecker(svc) // registers this service with the global health checker
	return svc
}

// HealthCheck returns the health status of the Transaction Management Service.
// It delegates the responsibility of executing the health check to the logic layer.
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

// GetTransactions returns a paginated list of transactions for the user.
// It extracts the user id from the session and retrieves the transactions using the logic layer.
func (svc transactionManagementService) GetTransactions(w http.ResponseWriter, r *http.Request) {
	sessionStruct := session.GetSession(r.Context())
	session, ok := sessionStruct.(model.SessionStruct)
	if !ok {
		response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrAssertUserid), nil)
		return
	}
	queryParams := r.URL.Query()
	limit, err := strconv.Atoi(queryParams.Get("limit"))
	if err != nil || limit == 0 {
		log.Info(fmt.Sprintf("setting default limit as %d as error: %+v, query: %s", 5, err, queryParams.Get("limit")))
		limit = 5
	}
	page, err := strconv.Atoi(queryParams.Get("page"))
	if err != nil || page == 0 {
		log.Info(fmt.Sprintf("setting default page as %d as error: %+v, query: %s", 1, err, queryParams.Get("page")))
		page = 1
	}
	resp := svc.logic.GetTransactions(session.UserId, limit, page)
	response.ToJson(w, resp.Status, resp.Message, resp.Data)
}

// NewTransaction creates a new transaction for the logged-in user using the data from the request body.
func (svc transactionManagementService) NewTransaction(w http.ResponseWriter, r *http.Request) {
	// Extract the user session from the request context.
	sessionStruct := session.GetSession(r.Context())
	session, ok := sessionStruct.(model.SessionStruct)
	if !ok {
		response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrAssertUserid), nil)
		return
	}
	// Parse the request body and validate the data.
	var newTransaction model.NewTransaction
	status, err := request.FromJson(r, &newTransaction)
	if err != nil {
		log.Error(err)
		response.ToJson(w, status, err.Error(), nil)
		return
	}
	// Set the user ID for the new transaction and pass it to the business logic.
	newTransaction.UserId = session.UserId
	resp := svc.logic.NewTransaction(newTransaction)
	// Return the response to the client in JSON format.
	response.ToJson(w, resp.Status, resp.Message, resp.Data)
}

// DownloadTransaction downloads a PDF file for a specific transaction ID belonging to the logged-in user.
func (svc transactionManagementService) DownloadTransaction(w http.ResponseWriter, r *http.Request) {
	// Extract the user session from the request context.
	sessionStruct := session.GetSession(r.Context())
	session, ok := sessionStruct.(model.SessionStruct)
	if !ok {
		response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrAssertUserid), nil)
		return
	}
	// Extract the transaction ID from the request parameters.
	vars := mux.Vars(r)
	if len(vars) == 0 {
		response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrGetTransaction), nil)
		return
	}
	// Download the PDF file for the given transaction ID and user session.
	resp := svc.logic.DownloadTransaction(vars["transaction_id"], session.Cookie)
	if resp.Status != http.StatusOK {
		response.ToJson(w, resp.Status, resp.Message, resp.Data)
		return
	}
	// Verify that the response data is a valid PDF byte slice and write it to the response writer.
	pdf, ok := resp.Data.([]byte)
	if !ok {
		response.ToJson(w, http.StatusBadRequest, codes.GetErr(codes.ErrAssertPdf), nil)
		return
	}
	_, err := w.Write(pdf)
	if err != nil {
		log.Error(err)
		response.ToJson(w, http.StatusInternalServerError, codes.GetErr(codes.ErrPdf), nil)
		return
	}
	// Set the appropriate headers for the PDF file download.
	w.Header().Set("Content-Disposition", "attachment; filename="+vars["transaction_id"]+".pdf")
	w.Header().Set("Content-Type", "application/pdf")
}
