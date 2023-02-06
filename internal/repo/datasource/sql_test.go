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
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fail()
	}
	svcConfig := config.SvcConfig{
		DbSvc: config.DbSvc{Db: db},
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
func TestSqlDs_Get(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() sqlDs
		cleanupFunc func()
		filter      map[string]interface{}
		validator   func([]model.Transaction, int, error)
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
				mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(`transaction_id`) FROM newTemp WHERE user_id = '1234' AND account_number = 1")).WillReturnError(nil).WillReturnRows(sqlmock.NewRows([]string{"count(transaction_id)"}).AddRow("1"))
				mock.ExpectQuery("SELECT transaction_id, account_number, user_id, amount, transfer_to, created_at, updated_at, status, type, comment FROM newTemp WHERE user_id = '1234' AND account_number = 1 ORDER BY created_at LIMIT 1 OFFSET 2 ;").WillReturnRows(sqlmock.NewRows([]string{"transaction_id", "account_number", "user_id", "amount", "transfer_to", "created_at", "updated_at", "status", "type", "comment"}).AddRow("0000-1111-2222-3333", 1, "4444-1111-2222-3333", 1000, 1234567890, time.Date(2023, time.December, 1, 1, 1, 1, 0, time.UTC), time.Date(2023, time.December, 1, 1, 1, 1, 0, time.UTC), "approved", "debit", "no comments"))
				return dB
			},
			validator: func(rows []model.Transaction, count int, err error) {
				temp := []model.Transaction{{
					TransactionId: "0000-1111-2222-3333",
					AccountNumber: 1,
					UserId:        "4444-1111-2222-3333",
					Amount:        1000,
					TransferTo:    1234567890,
					CreatedAt:     time.Date(2023, time.December, 1, 1, 1, 1, 0, time.UTC),
					UpdatedAt:     time.Date(2023, time.December, 1, 1, 1, 1, 0, time.UTC),
					Status:        "approved",
					Type:          "debit",
					Comment:       "no comments",
				}}
				if err != nil {
					t.Errorf("Want: %v, Got: %v", nil, err)
					return
				}
				if count != 1 {
					t.Errorf("Want: %v, Got: %v", 3, count)
					return
				}
				if !reflect.DeepEqual(rows, temp) {
					t.Errorf("Want: %v, Got: %v", temp, rows)
					return
				}
			},
		},
		{
			name:   "FAILURE::Get:: query error",
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
				mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(`transaction_id`) FROM newTemp WHERE userid = '1234'")).WillReturnError(nil).WillReturnRows(sqlmock.NewRows([]string{"transaction_id"}).AddRow("1").AddRow("2").AddRow("3"))
				mock.ExpectQuery("SELECT transaction_id, account_number, user_id, amount, transfer_to, created_at, updated_at, status, type, comment FROM newTemp WHERE userid = '1234' ORDER BY created_at LIMIT 1 OFFSET 2 ;").WillReturnError(errors.New("Unknown column"))
				return dB
			},
			validator: func(rows []model.Transaction, count int, err error) {
				if !strings.Contains(err.Error(), "Unknown column") {
					t.Errorf("Want: %v, Got: %v", "Unknown column", err)
				}
			},
		},
		{
			name:   "FAILURE::Get:: query error::2",
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
				mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(`transaction_id`) FROM newTemp WHERE userid = '1234'")).WillReturnError(errors.New("query error")).WillReturnRows(sqlmock.NewRows([]string{"transaction_id"}).AddRow("1").AddRow("2").AddRow("3"))
				mock.ExpectQuery("SELECT transaction_id, account_number, user_id, amount, transfer_to, created_at, updated_at, status, type, comment FROM newTemp WHERE user_id = '1234' ORDER BY created_at LIMIT 1 OFFSET 2 ;").WillReturnError(errors.New("Unknown column"))
				return dB
			},
			validator: func(rows []model.Transaction, count int, err error) {
				if !strings.Contains(err.Error(), "query error") {
					t.Errorf("Want: %v, Got: %v", "query error", err)
				}
			},
		},
		{
			name: "failure::Get::scan error", //scan should return an error
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
				mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(`transaction_id`) FROM newTemp WHERE user_id = '1234'")).WillReturnError(nil).WillReturnRows(sqlmock.NewRows([]string{"transaction_id"}).AddRow("1").AddRow("2").AddRow("3"))
				mock.ExpectQuery("SELECT transaction_id, account_number, user_id, amount, transfer_to, created_at, updated_at, status, type, comment FROM newTemp WHERE user_id = '1234' ORDER BY created_at LIMIT 1 OFFSET 2 ;").WillReturnRows(sqlmock.NewRows([]string{"transaction_id", "account_number", "user_id", "amount", "transfer_to", "created_at", "updated_at", "status", "type", "comment"}).AddRow(true, 1, "123", 1000, 1234567890, time.Now(), "abc", "approved", "debit", "no comments"))
				return dB
			},
			validator: func(rows []model.Transaction, count int, err error) {
				if !strings.Contains(err.Error(), "sql: Scan error on column") {
					t.Errorf("Want: %v, Got: %v", "sql: Scan error on column", err.Error())
				}
			},
		},
		{
			name:   "FAILURE::Get:: scan error:: 2",
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
				mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(`transaction_id`) FROM newTemp WHERE userid = '1234'")).WillReturnError(nil).WillReturnRows(sqlmock.NewRows([]string{"count(transaction_id)"}).AddRow(true))
				mock.ExpectQuery("SELECT transaction_id, account_number, user_id, amount, transfer_to, created_at, updated_at, status, type, comment FROM newTemp WHERE userid = '1234' ORDER BY created_at LIMIT 1 OFFSET 2 ;").WillReturnRows(sqlmock.NewRows([]string{"1"}))
				return dB
			},
			validator: func(rows []model.Transaction, count int, err error) {
				if !strings.Contains(err.Error(), "sql: Scan error") {
					t.Errorf("Want: %v, Got: %v", "sql: Scan error", err)
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
			rows, count, err := db.Get(tt.filter, 1, 2)

			// STEP 3: validation of output
			if tt.validator != nil {
				tt.validator(rows, count, err)
			}

			// STEP 4: clean up/remove up all instances for the specific test case
			if tt.cleanupFunc != nil {
				tt.cleanupFunc()
			}
		})
	}
}

//
func TestSqlDs_Insert(t *testing.T) {
	// table driven tests
	tests := []struct {
		name        string
		tableName   string
		data        model.Transaction
		setupFunc   func() (sqlDs, sqlmock.Sqlmock)
		cleanupFunc func()
		filter      map[string]interface{}
		validator   func(sqlmock.Sqlmock, error)
	}{
		{
			name: "SUCCESS:: Insert Transaction",
			data: model.Transaction{
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
				m := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO newTemp(user_id, transaction_id, account_number, amount, transfer_to, status, type, comment) VALUES(?,?,?,?,?,?,?,?)")).WithArgs("1", "1234", 1, 1000.00, 2, "approved", "debit", "abcd")
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
			data: model.Transaction{
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
