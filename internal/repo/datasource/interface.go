package datasource

import "github.com/vatsal278/TransactionManagementService/internal/model"

//go:generate mockgen --build_flags=--mod=mod --destination=./../../../pkg/mock/mock_datasource.go --package=mock github.com/vatsal278/TransactionManagementService/internal/repo/datasource DataSourceI

type DataSourceI interface {
	HealthCheck() bool
	Get(map[string]interface{}, int, int) ([]model.GetTransaction, int, error)
	Insert(user model.NewTransaction) error
	//Update(filterSet map[string]interface{}, filterWhere map[string]interface{}) error
}
