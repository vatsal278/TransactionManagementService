package datasource

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	svcCfg "github.com/vatsal278/AccountManagmentSvc/internal/config"
	"github.com/vatsal278/AccountManagmentSvc/internal/model"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestSqlDs_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing due to unavailability of testing environment")
	}
	//include a failure case
	dbcfg := svcCfg.DbCfg{
		Port:      "9085",
		Host:      "localhost",
		Driver:    "mysql",
		User:      "root",
		Pass:      "pass",
		DbName:    "useracc",
		TableName: "newTemp",
	}
	dataBase := svcCfg.Connect(dbcfg, dbcfg.TableName)
	svcConfig := svcCfg.SvcConfig{
		DbSvc: svcCfg.DbSvc{Db: dataBase},
	}
	dB := NewSql(svcConfig.DbSvc, "newTemp")

	tests := []struct {
		name        string
		setupFunc   func()
		cleanupFunc func()
		filter      map[string]interface{}
		validator   func(bool)
		dbInterface DataSourceI
	}{
		{
			name: "SUCCESS::Health check",
			validator: func(res bool) {
				if res != true {
					t.Errorf("Want: %v, Got: %v", true, res)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			res := dB.HealthCheck()

			if tt.validator != nil {
				tt.validator(res)
			}
		})
	}
}
func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() sqlDs
		cleanupFunc func()
		filter      map[string]interface{}
		validator   func([]model.Account, error)
		dbInterface DataSourceI
	}{
		{
			name: "SUCCESS::Get",
			filter: map[string]interface{}{
				"user_id": "1234",
			},
			setupFunc: func() sqlDs {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fail()
				}
				dB := sqlDs{
					sqlSvc: db,
					table:  "newTemp",
				}
				mock.ExpectQuery("SELECT user_id, account_number, income, spends, created_on, updated_on, active_services, inactive_services FROM newTemp WHERE user_id = '1234' ORDER BY account_number;").WillReturnRows(sqlmock.NewRows([]string{"user_id", "account_number", "income", "spends", "created_on", "updated_on", "active_services", "inactive_services"}).AddRow("1234", 1, 0.00, 0.00, time.Now(), time.Now(), &model.Svc{"1": {}}, &model.Svc{"1": {}}).RowError(1, errors.New("")))
				return dB
			},
			validator: func(rows []model.Account, err error) {
				temp := model.Account{
					Id:               "1234",
					AccountNumber:    1,
					CreatedOn:        time.Now(),
					ActiveServices:   &model.Svc{"1": {}},
					InactiveServices: &model.Svc{"1": {}},
				}
				if err != nil {
					t.Errorf("Want: %v, Got: %v", nil, err)
				}
				if !reflect.DeepEqual(rows[0].Id, temp.Id) {
					t.Errorf("Want: %v, Got: %v", temp.Id, rows[0].Id)
				}
				if !reflect.DeepEqual(rows[0].AccountNumber, temp.AccountNumber) {
					t.Errorf("Want: %v, Got: %v", temp.AccountNumber, rows[0].AccountNumber)
				}
				if !reflect.DeepEqual(rows[0].Income, temp.Income) {
					t.Errorf("Want: %v, Got: %v", temp.Income, rows[0].Income)
				}
				if !reflect.DeepEqual(rows[0].Spends, temp.Spends) {
					t.Errorf("Want: %v, Got: %v", temp.Spends, rows[0].Spends)
				}
				if !reflect.DeepEqual(rows[0].ActiveServices, temp.ActiveServices) {
					t.Errorf("Want: %v, Got: %v", temp.ActiveServices, rows[0].ActiveServices)
				}
				if !reflect.DeepEqual(rows[0].InactiveServices, temp.InactiveServices) {
					t.Errorf("Want: %v, Got: %v", temp.InactiveServices, rows[0].InactiveServices)
				}

			},
		},
		{
			name: "SUCCESS::Get:: multiple articles",
			filter: map[string]interface{}{
				"user_id": "1234",
			},
			setupFunc: func() sqlDs {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fail()
				}
				dB := sqlDs{
					sqlSvc: db,
					table:  "newTemp",
				}
				mock.ExpectQuery("SELECT user_id, account_number, income, spends, created_on, updated_on, active_services, inactive_services FROM newTemp WHERE user_id = '1234' ORDER BY account_number;").WillReturnRows(sqlmock.NewRows([]string{"user_id", "account_number", "income", "spends", "created_on", "updated_on", "active_services", "inactive_services"}).AddRow("1234", 1, 0.00, 0.00, time.Now(), time.Now(), &model.Svc{"1": {}}, &model.Svc{"1": {}}).AddRow("12345", 1, 0.00, 0.00, time.Now(), time.Now(), &model.Svc{"1": {}}, &model.Svc{"1": {}}))
				return dB
			},
			validator: func(rows []model.Account, err error) {
				temp := model.Account{
					Id:               "1234",
					AccountNumber:    1,
					InactiveServices: &model.Svc{"1": {}},
				}
				if err != nil {
					t.Errorf("Want: %v, Got: %v", nil, err)
				}
				if !reflect.DeepEqual(rows[0].Id, temp.Id) {
					t.Errorf("Want: %v, Got: %v", temp.Id, rows[0].Id)
				}
				if !reflect.DeepEqual(rows[0].AccountNumber, temp.AccountNumber) {
					t.Errorf("Want: %v, Got: %v", temp.AccountNumber, rows[0].AccountNumber)
				}
				if !reflect.DeepEqual(rows[0].Income, temp.Income) {
					t.Errorf("Want: %v, Got: %v", temp.Income, rows[0].Income)
				}
				if !reflect.DeepEqual(rows[0].Spends, temp.Spends) {
					t.Errorf("Want: %v, Got: %v", temp.Spends, rows[0].Spends)
				}
				if !reflect.DeepEqual(rows[0].InactiveServices, temp.InactiveServices) {
					t.Errorf("Want: %v, Got: %v", temp.InactiveServices, rows[0].InactiveServices)
				}
				if !reflect.DeepEqual(rows[0].ActiveServices, temp.InactiveServices) {
					t.Errorf("Want: %v, Got: %v", temp.InactiveServices, rows[0].InactiveServices)
				}
			},
		},
		{
			name: "SUCCESS::Get::no user found",
			filter: map[string]interface{}{
				"account_number": 1,
			},
			setupFunc: func() sqlDs {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fail()
				}
				dB := sqlDs{
					sqlSvc: db,
					table:  "newTemp",
				}
				mock.ExpectQuery("SELECT user_id, account_number, income, spends, created_on, updated_on, active_services, inactive_services FROM newTemp WHERE user_id = '1234' ORDER BY account_number;").WillReturnRows(sqlmock.NewRows([]string{"user_id", "account_number", "income", "spends", "created_on", "updated_on", "active_services", "inactive_services"}))
				return dB
			},
			validator: func(rows []model.Account, err error) {
				if len(rows) != 0 {
					t.Errorf("Want: %v, Got: %v", 0, len(rows))
				}
			},
		},
		{
			name: "failure::Get::scan error", //scan should return an error
			filter: map[string]interface{}{
				"user_id": "12345",
			},
			setupFunc: func() sqlDs {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fail()
				}
				dB := sqlDs{
					sqlSvc: db,
					table:  "newTemp",
				}
				mock.ExpectQuery("SELECT user_id, account_number, income, spends, created_on, updated_on, active_services, inactive_services FROM newTemp WHERE user_id = '12345' ORDER BY account_number;").WillReturnRows(sqlmock.NewRows([]string{"user_id", "account_number", "income", "spends", "created_on", "updated_on", "active_services", "inactive_services"}).AddRow("12345	", 1, 0.00, "abc", time.Now(), time.Now(), &model.Svc{"1": {}}, &model.Svc{"1": {}}))
				return dB
			},
			validator: func(rows []model.Account, err error) {
				if !strings.Contains(err.Error(), "sql: Scan error on column") {
					t.Errorf("Want: %v, Got: %v", "sql: Scan error on column", err.Error())
				}
			},
		},
		{
			name:   "FAILURE:: query error",
			filter: map[string]interface{}{"userid": "1234"},
			setupFunc: func() sqlDs {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fail()
				}
				dB := sqlDs{
					sqlSvc: db,
					table:  "newTemp",
				}
				mock.ExpectQuery("SELECT user_id, account_number, income, spends, created_on, updated_on, active_services, inactive_services FROM newTemp WHERE userid = '1234' ORDER BY account_number;").WillReturnError(errors.New("Unknown column"))
				return dB
			},
			validator: func(rows []model.Account, err error) {
				if !strings.Contains(err.Error(), "Unknown column") {
					t.Errorf("Want: %v, Got: %v", "Unknown column", err)
				}
			},
		},
	}

	// to execute the tests in the table
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// STEP 1: seting up all instances for the specific test case
			db := tt.setupFunc()
			// STEP 2: call the test function
			rows, err := db.Get(tt.filter)

			// STEP 3: validation of output
			if tt.validator != nil {
				tt.validator(rows, err)
			}

			// STEP 4: clean up/remove up all instances for the specific test case
			if tt.cleanupFunc != nil {
				tt.cleanupFunc()
			}
		})
	}
}

//
func TestInsert(t *testing.T) {
	// table driven tests
	tests := []struct {
		name        string
		tableName   string
		data        model.Account
		setupFunc   func() (sqlDs, sqlmock.Sqlmock)
		cleanupFunc func()
		filter      map[string]interface{}
		validator   func(sqlmock.Sqlmock, error)
	}{
		{
			name: "SUCCESS:: Insert Article",
			data: model.Account{
				Id:               "1",
				ActiveServices:   &model.Svc{"1": {}},
				InactiveServices: &model.Svc{},
			},
			setupFunc: func() (sqlDs, sqlmock.Sqlmock) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fail()
				}
				dB := sqlDs{
					sqlSvc: db,
					table:  "newTemp",
				}
				m := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO newTemp(user_id, active_services, inactive_services) VALUES(?,?,?)")).WithArgs("1", &model.Svc{"1": {}}, &model.Svc{})
				m.WillReturnError(nil)
				m.WillReturnResult(sqlmock.NewResult(1, 1))
				return dB, mock
			},
			validator: func(mock sqlmock.Sqlmock, err error) {
				if mock.ExpectationsWereMet() != nil {
					t.Errorf("Want: %v, Got: %v", nil, err.Error())
					return
				}
				if err != nil {
					t.Errorf("Want: %v, Got: %v", nil, err.Error())
					return
				}
			},
		},
		{
			name: "SUCCESS:: Insert Article:: Insert Article when data already present",
			data: model.Account{
				Id:               "2",
				ActiveServices:   &model.Svc{"1": {}},
				InactiveServices: &model.Svc{"2": {}},
			},
			setupFunc: func() (sqlDs, sqlmock.Sqlmock) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fail()
				}
				dB := sqlDs{
					sqlSvc: db,
					table:  "newTemp",
				}
				m := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO newTemp(user_id, active_services, inactive_services) VALUES(?,?,?)")).WithArgs("2", &model.Svc{"1": {}}, &model.Svc{"2": {}})
				m.WillReturnError(nil)
				m.WillReturnResult(sqlmock.NewResult(2, 1))
				return dB, mock
			},
			validator: func(mock sqlmock.Sqlmock, err error) {
				if mock.ExpectationsWereMet() != nil {
					t.Errorf("Want: %v, Got: %v", nil, err.Error())
					return
				}
				if err != nil {
					t.Errorf("Want: %v, Got: %v", nil, err.Error())
					return
				}
			},
		},
		{
			name: "FAILURE:: insert :: sql error",
			data: model.Account{
				Id:               "2",
				ActiveServices:   &model.Svc{"1": {}},
				InactiveServices: nil,
			},
			setupFunc: func() (sqlDs, sqlmock.Sqlmock) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fail()
				}
				dB := sqlDs{
					sqlSvc: db,
					table:  "newTemp",
				}
				m := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO newTemp(user_id, active_services, inactive_services) VALUES(?,?,?)")).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg())
				m.WillReturnError(errors.New("sql error"))
				m.WillReturnResult(sqlmock.NewResult(0, 0))
				return dB, mock
			},
			validator: func(mock sqlmock.Sqlmock, err error) {
				if err.Error() != errors.New("sql error").Error() {
					t.Errorf("Want: %v, Got: %v", "sql error", err.Error())
					return
				}
			},
		},
	}
	// to execute the tests in the table
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.setupFunc()
			// STEP 2: call the test function
			err := db.Insert(tt.data)
			// STEP 3: validation of output
			if tt.validator != nil {
				tt.validator(mock, err)
			}
			// STEP 4: clean up/remove up all instances for the specific test case
			if tt.cleanupFunc != nil {
				tt.cleanupFunc()
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name        string
		tableName   string
		dataSet     map[string]interface{}
		dataWhere   map[string]interface{}
		setupFunc   func() (sqlDs, sqlmock.Sqlmock)
		cleanupFunc func()
		filter      map[string]interface{}
		validator   func(error, sqlmock.Sqlmock)
	}{
		{
			name:      "SUCCESS:: Update",
			dataSet:   map[string]interface{}{"income": model.ColumnUpdate{UpdateSet: "income+100"}},
			dataWhere: map[string]interface{}{"user_id": "100"},
			setupFunc: func() (sqlDs, sqlmock.Sqlmock) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fail()
				}
				dB := sqlDs{
					sqlSvc: db,
					table:  "newTemp",
				}
				mock.ExpectExec(regexp.QuoteMeta("UPDATE newTemp  SET income = income+100 WHERE user_id = '100' ;")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				return dB, mock
			},
			validator: func(err error, mock sqlmock.Sqlmock) {
				if mock.ExpectationsWereMet() != nil {
					t.Errorf("Want: %v, Got: %v", nil, err.Error())
					return
				}
				if err != nil {
					t.Errorf("Want: %v, Got: %v", nil, err.Error())
					return
				}
			},
		},
		{
			name:      "Success:: Update:: removing and inserting",
			dataSet:   map[string]interface{}{"active_services": model.ColumnUpdate{UpdateSet: "JSON_INSERT(active_services, '$.\"1\"', JSON_OBJECT())"}, "inactive_services": model.ColumnUpdate{UpdateSet: "JSON_REMOVE(inactive_services, '$.\"1\"')"}},
			dataWhere: map[string]interface{}{"user_id": "1233"},
			setupFunc: func() (sqlDs, sqlmock.Sqlmock) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fail()
				}
				dB := sqlDs{
					sqlSvc: db,
					table:  "newTemp",
				}
				mock.ExpectExec(regexp.QuoteMeta("UPDATE newTemp SET active_services = JSON_INSERT(active_services, '$.\"1\"', JSON_OBJECT()) , inactive_services = JSON_REMOVE(inactive_services, '$.\"1\"') WHERE user_id = '1233' ;")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
				return dB, mock
			},
			validator: func(err error, mock sqlmock.Sqlmock) {
				if mock.ExpectationsWereMet() != nil {
					t.Errorf("Want: %v, Got: %v", nil, err.Error())
					return
				}
				if err != nil {
					t.Errorf("Want: %v, Got: %v", nil, err.Error())
					return
				}
			},
		},
		{
			name:      "Failure:: Update::",
			dataSet:   map[string]interface{}{"active_services": model.ColumnUpdate{UpdateSet: "JSON_INSERT(active_services, '$.\"1\"', JSON_OBJECT())"}},
			dataWhere: map[string]interface{}{"abc": "1233"},
			setupFunc: func() (sqlDs, sqlmock.Sqlmock) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fail()
				}
				dB := sqlDs{
					sqlSvc: db,
					table:  "newTemp",
				}
				mock.ExpectExec(regexp.QuoteMeta("UPDATE newTemp  SET active_services = JSON_INSERT(active_services, '$.\"1\"', JSON_OBJECT()) WHERE abc = '1233' ;")).WillReturnError(errors.New("unknown column abc")).WillReturnResult(sqlmock.NewResult(1, 1))
				return dB, mock
			},
			validator: func(err error, mock sqlmock.Sqlmock) {
				if mock.ExpectationsWereMet() != nil {
					t.Errorf("Want: %v, Got: %v", nil, err.Error())
					return
				}
				if err.Error() != errors.New("unknown column abc").Error() {
					t.Errorf("Want: %v, Got: %v", "unknown column abc", err.Error())
					return
				}
			},
		},
	}
	// to execute the tests in the table
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.setupFunc()
			// STEP 2: call the test function
			err := db.Update(tt.dataSet, tt.dataWhere)

			// STEP 3: validation of output
			if tt.validator != nil {
				tt.validator(err, mock)
			}

			// STEP 4: clean up/remove up all instances for the specific test case
			if tt.cleanupFunc != nil {
				tt.cleanupFunc()
			}
		})
	}
}
