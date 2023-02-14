package router

import (
	"net/http"

	"github.com/PereRohit/util/constant"
	"github.com/PereRohit/util/middleware"
	"github.com/gorilla/mux"

	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/handler"
	middleware2 "github.com/vatsal278/TransactionManagementService/internal/middleware"
	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
)

// Register creates a new mux.Router and attaches all the necessary routes for the service
func Register(svcCfg *config.SvcConfig) *mux.Router {
	m := mux.NewRouter()

	// group all routes for specific version. e.g.: /v1
	if svcCfg.ServiceRouteVersion != "" {
		m = m.PathPrefix("/" + svcCfg.ServiceRouteVersion).Subrouter()
	}

	m.StrictSlash(true)

	// middleware for request hijacking and panic recovery
	m.Use(middleware.RequestHijacker)
	m.Use(middleware.RecoverPanic)

	// handler for common service routes
	commons := handler.NewCommonSvc()
	m.HandleFunc(constant.HealthRoute, commons.HealthCheck).Methods(http.MethodGet)
	m.NotFoundHandler = http.HandlerFunc(commons.RouteNotFound)
	m.MethodNotAllowedHandler = http.HandlerFunc(commons.MethodNotAllowed)

	// attach routes for services below
	m = attachTransactionManagementServiceRoutes(m, svcCfg)

	return m
}

// attachTransactionManagementServiceRoutes attaches all the routes for the TransactionManagementService
func attachTransactionManagementServiceRoutes(m *mux.Router, svcCfg *config.SvcConfig) *mux.Router {
	// create new datasource for the TransactionManagementService
	dataSource := datasource.NewSql(svcCfg.DbSvc, svcCfg.Cfg.DataBase.TableName)

	// create new middleware for the TransactionManagementService
	middleware := middleware2.NewTransactionMgmtMiddleware(svcCfg)

	// create new handler for the TransactionManagementService
	svc := handler.NewTransactionManagementService(dataSource, svcCfg.ExternalService)

	// create new subrouter for the new transaction route
	router := m.PathPrefix("").Subrouter()
	router.HandleFunc("", svc.NewTransaction).Methods(http.MethodPost)
	router.HandleFunc("/download/{transaction_id}", svc.DownloadTransaction).Methods(http.MethodGet)

	// attach middleware to the new transaction route
	router.Use(middleware.ExtractUser)

	// create new subrouter for the get transactions route
	router2 := m.PathPrefix("").Subrouter()
	router2.HandleFunc("", svc.GetTransactions).Methods(http.MethodGet)

	// attach middleware to the get transactions route
	router2.Use(middleware.ExtractUser)
	router2.Use(middleware.Cacher(true))

	return m
}
