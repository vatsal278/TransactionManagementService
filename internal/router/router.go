package router

import (
	middleware2 "github.com/vatsal278/TransactionManagementService/internal/middleware"
	"net/http"

	"github.com/PereRohit/util/constant"
	"github.com/PereRohit/util/middleware"
	"github.com/gorilla/mux"

	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/handler"
	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
)

func Register(svcCfg *config.SvcConfig) *mux.Router {
	m := mux.NewRouter()

	// group all routes for specific version. e.g.: /v1
	if svcCfg.ServiceRouteVersion != "" {
		m = m.PathPrefix("/" + svcCfg.ServiceRouteVersion).Subrouter()
	}

	m.StrictSlash(true)
	m.Use(middleware.RequestHijacker)
	m.Use(middleware.RecoverPanic)

	commons := handler.NewCommonSvc()
	m.HandleFunc(constant.HealthRoute, commons.HealthCheck).Methods(http.MethodGet)
	m.NotFoundHandler = http.HandlerFunc(commons.RouteNotFound)
	m.MethodNotAllowedHandler = http.HandlerFunc(commons.MethodNotAllowed)

	// attach routes for services below
	m = attachTransactionManagementServiceRoutes(m, svcCfg)

	return m
}

func attachTransactionManagementServiceRoutes(m *mux.Router, svcCfg *config.SvcConfig) *mux.Router {
	dataSource := datasource.NewSql(svcCfg.DbSvc, svcCfg.Cfg.DataBase.TableName)
	middleware := middleware2.NewTransactionMgmtMiddleware(svcCfg)
	svc := handler.NewTransactionManagementService(dataSource, svcCfg.UtilService)

	router := m.PathPrefix("").Subrouter()
	router.HandleFunc("", svc.NewTransaction).Methods(http.MethodPost)
	router.HandleFunc("/download/transaction_id", svc.NewTransaction).Methods(http.MethodGet)
	router.Use(middleware.ExtractUser)

	router2 := m.PathPrefix("").Subrouter()
	router2.HandleFunc("", svc.GetTransactions).Methods(http.MethodGet)
	router2.Use(middleware.ExtractUser)
	router2.Use(middleware.Cacher(true))

	return m

}
