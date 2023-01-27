package datasource

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/model"
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
	dbcfg := config.DbCfg{
		Port:      "9085",
		Host:      "localhost",
		Driver:    "mysql",
		User:      "root",
		Pass:      "pass",
		DbName:    "useracc",
		TableName: "newTemp",
	}
	dataBase := config.Connect(dbcfg, dbcfg.TableName)
	svcConfig := config.SvcConfig{
		DbSvc: config.DbSvc{Db: dataBase},
	}
	dB := NewSql(config.DbSvc(svcConfig.DbSvc), "newTemp")

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
		validator   func([]model.GetTransaction, error)
		dbInterface DataSourceI
	}{
		{
			name: "SUCCESS::Get",
			filter: map[string]interface{}{
				"user_id":        "1234",
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
				mock.ExpectQuery("SELECT transaction_id, account_number, amount, transfer_to, created_at, updated_at, status, type, comment FROM newTemp WHERE user_id = '1234' AND account_number = 1 ORDER BY created_at LIMIT 1 OFFSET 2 ;").WillReturnRows(sqlmock.NewRows([]string{"transaction_id", "account_number", "amount", "transfer_to", "created_at", "updated_at", "status", "type", "comment"}).AddRow("0000-1111-2222-3333", 1, 1000, 1234567890, time.Now(), time.Now(), "approved", "debit", "no comments"))
				t.Log(mock.ExpectationsWereMet())
				return dB
			},
			validator: func(rows []model.GetTransaction, err error) {
				temp := model.GetTransaction{
					TransactionId: "0000-1111-2222-3333",
					AccountNumber: 1,
					Amount:        1000,
					TransferTo:    1234567890,
					Status:        "approved",
					Type:          "debit",
					Comment:       "no comments",
				}
				if err != nil {
					t.Errorf("Want: %v, Got: %v", nil, err)
				}
				if !reflect.DeepEqual(rows[0].TransactionId, temp.TransactionId) {
					t.Errorf("Want: %v, Got: %v", temp.TransactionId, rows[0].TransactionId)
				}
				if !reflect.DeepEqual(rows[0].AccountNumber, temp.AccountNumber) {
					t.Errorf("Want: %v, Got: %v", temp.AccountNumber, rows[0].AccountNumber)
				}
				if !reflect.DeepEqual(rows[0].Amount, temp.Amount) {
					t.Errorf("Want: %v, Got: %v", temp.Amount, rows[0].Amount)
				}
				if !reflect.DeepEqual(rows[0].TransferTo, temp.TransferTo) {
					t.Errorf("Want: %v, Got: %v", temp.TransferTo, rows[0].TransferTo)
				}
				if !reflect.DeepEqual(rows[0].Type, temp.Type) {
					t.Errorf("Want: %v, Got: %v", temp.Type, rows[0].Type)
				}
				if !reflect.DeepEqual(rows[0].Status, temp.Status) {
					t.Errorf("Want: %v, Got: %v", temp.Status, rows[0].Status)
				}
				if !reflect.DeepEqual(rows[0].Comment, temp.Comment) {
					t.Errorf("Want: %v, Got: %v", temp.Comment, rows[0].Comment)
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
				mock.ExpectQuery("SELECT transaction_id, account_number, amount, transfer_to, created_at, updated_at, status, type, comment FROM newTemp WHERE user_id = '12345' ORDER BY created_at LIMIT 1 OFFSET 2 ;").WillReturnRows(sqlmock.NewRows([]string{"transaction_id", "account_number", "amount", "transfer_to", "created_at", "updated_at", "status", "type", "comment"}).AddRow(true, 1, 1000, 1234567890, time.Now(), "abc", "approved", "debit", "no comments"))
				return dB
			},
			validator: func(rows []model.GetTransaction, err error) {
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
				mock.ExpectQuery("SELECT transaction_id, account_number, amount, transfer_to, created_at, updated_at, status, type, comment FROM newTemp WHERE userid = '1234' ORDER BY created_at LIMIT 1 OFFSET 2 ;").WillReturnError(errors.New("Unknown column"))
				return dB
			},
			validator: func(rows []model.GetTransaction, err error) {
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
			rows, err := db.Get(tt.filter, 1, 2)

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
		data        model.NewTransaction
		setupFunc   func() (sqlDs, sqlmock.Sqlmock)
		cleanupFunc func()
		filter      map[string]interface{}
		validator   func(sqlmock.Sqlmock, error)
	}{
		{
			name: "SUCCESS:: Insert Article",
			data: model.NewTransaction{
				UserId:        "1",
				AccountNumber: 1,
				TransactionId: "1234",
				Amount:        1000,
				TransferTo:    2,
				Status:        "approved",
				Type:          "debit",
				Comment:       "abcd",
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
				m := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO newTemp(user_id, transaction_id, account_number, amount, transfer_to, status, type, comment) VALUES(?,?,?,?,?,?,?,?)")).WithArgs("1", "1234", 1, 1000, 2, "approved", "debit", "abcd")
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
			name: "FAILURE:: insert :: sql error",
			data: model.NewTransaction{
				UserId:        "1",
				AccountNumber: 1,
				TransactionId: "1234",
				Amount:        1000,
				TransferTo:    2,
				Status:        "approved",
				Type:          "debit",
				Comment:       "abcd",
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
				m := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO newTemp(user_id, transaction_id, account_number, amount, transfer_to, status, type, comment) VALUES(?,?,?,?,?,?,?,?)")).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg())
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

//func TestUpdate(t *testing.T) {
//	tests := []struct {
//		name        string
//		tableName   string
//		dataSet     map[string]interface{}
//		dataWhere   map[string]interface{}
//		setupFunc   func() (sqlDs, sqlmock.Sqlmock)
//		cleanupFunc func()
//		filter      map[string]interface{}
//		validator   func(error, sqlmock.Sqlmock)
//	}{
//		{
//			name:      "SUCCESS:: Update",
//			dataSet:   map[string]interface{}{"income": model.ColumnUpdate{UpdateSet: "income+100"}},
//			dataWhere: map[string]interface{}{"user_id": "100"},
//			setupFunc: func() (sqlDs, sqlmock.Sqlmock) {
//				db, mock, err := sqlmock.New()
//				if err != nil {
//					t.Fail()
//				}
//				dB := sqlDs{
//					sqlSvc: db,
//					table:  "newTemp",
//				}
//				mock.ExpectExec(regexp.QuoteMeta("UPDATE newTemp  SET income = income+100 WHERE user_id = '100' ;")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
//				return dB, mock
//			},
//			validator: func(err error, mock sqlmock.Sqlmock) {
//				if mock.ExpectationsWereMet() != nil {
//					t.Errorf("Want: %v, Got: %v", nil, err.Error())
//					return
//				}
//				if err != nil {
//					t.Errorf("Want: %v, Got: %v", nil, err.Error())
//					return
//				}
//			},
//		},
//		{
//			name:      "Success:: Update:: removing and inserting",
//			dataSet:   map[string]interface{}{"active_services": model.ColumnUpdate{UpdateSet: "JSON_INSERT(active_services, '$.\"1\"', JSON_OBJECT())"}, "inactive_services": model.ColumnUpdate{UpdateSet: "JSON_REMOVE(inactive_services, '$.\"1\"')"}},
//			dataWhere: map[string]interface{}{"user_id": "1233"},
//			setupFunc: func() (sqlDs, sqlmock.Sqlmock) {
//				db, mock, err := sqlmock.New()
//				if err != nil {
//					t.Fail()
//				}
//				dB := sqlDs{
//					sqlSvc: db,
//					table:  "newTemp",
//				}
//				mock.ExpectExec(regexp.QuoteMeta("UPDATE newTemp SET active_services = JSON_INSERT(active_services, '$.\"1\"', JSON_OBJECT()) , inactive_services = JSON_REMOVE(inactive_services, '$.\"1\"') WHERE user_id = '1233' ;")).WillReturnError(nil).WillReturnResult(sqlmock.NewResult(1, 1))
//				return dB, mock
//			},
//			validator: func(err error, mock sqlmock.Sqlmock) {
//				if mock.ExpectationsWereMet() != nil {
//					t.Errorf("Want: %v, Got: %v", nil, err.Error())
//					return
//				}
//				if err != nil {
//					t.Errorf("Want: %v, Got: %v", nil, err.Error())
//					return
//				}
//			},
//		},
//		{
//			name:      "Failure:: Update::",
//			dataSet:   map[string]interface{}{"active_services": model.ColumnUpdate{UpdateSet: "JSON_INSERT(active_services, '$.\"1\"', JSON_OBJECT())"}},
//			dataWhere: map[string]interface{}{"abc": "1233"},
//			setupFunc: func() (sqlDs, sqlmock.Sqlmock) {
//				db, mock, err := sqlmock.New()
//				if err != nil {
//					t.Fail()
//				}
//				dB := sqlDs{
//					sqlSvc: db,
//					table:  "newTemp",
//				}
//				mock.ExpectExec(regexp.QuoteMeta("UPDATE newTemp  SET active_services = JSON_INSERT(active_services, '$.\"1\"', JSON_OBJECT()) WHERE abc = '1233' ;")).WillReturnError(errors.New("unknown column abc")).WillReturnResult(sqlmock.NewResult(1, 1))
//				return dB, mock
//			},
//			validator: func(err error, mock sqlmock.Sqlmock) {
//				if mock.ExpectationsWereMet() != nil {
//					t.Errorf("Want: %v, Got: %v", nil, err.Error())
//					return
//				}
//				if err.Error() != errors.New("unknown column abc").Error() {
//					t.Errorf("Want: %v, Got: %v", "unknown column abc", err.Error())
//					return
//				}
//			},
//		},
//	}
//	// to execute the tests in the table
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			db, mock := tt.setupFunc()
//			// STEP 2: call the test function
//			err := db.Update(tt.dataSet, tt.dataWhere)
//
//			// STEP 3: validation of output
//			if tt.validator != nil {
//				tt.validator(err, mock)
//			}
//
//			// STEP 4: clean up/remove up all instances for the specific test case
//			if tt.cleanupFunc != nil {
//				tt.cleanupFunc()
//			}
//		})
//	}
//}
