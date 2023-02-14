package datasource

import (
	"database/sql"
	"fmt"
	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/model"
	"strings"
)

type sqlDs struct {
	sqlSvc *sql.DB
	table  string
}

// NewSql creates a new instance of sqlDs with a given database service and table name.
func NewSql(dbSvc config.DbSvc, tableName string) DataSourceI {
	return &sqlDs{
		sqlSvc: dbSvc.Db,
		table:  tableName,
	}
}

// queryFromMap returns a string containing SQL queries generated from a given map of filters and a join separator.
func queryFromMap(d map[string]interface{}, join string) string {
	var (
		q string
		f []string
	)
	for k, v := range d {
		switch v.(type) {
		case string:
			f = append(f, fmt.Sprintf(`%s = '%s'`, k, v))
		default:
			f = append(f, fmt.Sprintf(`%s = %v`, k, v))
		}
	}
	if len(f) > 0 {
		q = fmt.Sprintf(`%s`, strings.Join(f, ` `+join+` `))
	}
	return q
}

// HealthCheck checks the health of the database service.
func (d sqlDs) HealthCheck() bool {
	err := d.sqlSvc.Ping()
	return err == nil
}

// Get retrieves transactions from the database service based on a given set of filters, limit, and offset.
func (d sqlDs) Get(filter map[string]interface{}, limit int, offset int) ([]model.Transaction, int, error) {
	var transaction model.Transaction
	var transactions []model.Transaction
	var count int
	q := fmt.Sprintf("SELECT transaction_id, account_number, user_id, amount, transfer_to, created_at, updated_at, status, type, comment FROM %s", d.table)
	whereQuery := queryFromMap(filter, " AND ")
	if whereQuery != "" {
		whereQuery = " WHERE " + whereQuery
		q += whereQuery
	}
	queryCount := fmt.Sprintf("SELECT COUNT(`transaction_id`) FROM %s %s", d.table, whereQuery)
	rowsCount, err := d.sqlSvc.Query(queryCount)
	if err != nil {
		return nil, 0, err
	}
	for rowsCount.Next() {
		err = rowsCount.Scan(&count)
		if err != nil {
			return nil, 0, err
		}
	}
	r := " ORDER BY created_at ;"
	if limit > 0 {
		r = fmt.Sprintf(" ORDER BY created_at LIMIT %d OFFSET %d ;", limit, offset)
	}
	q += r
	rows, err := d.sqlSvc.Query(q)
	if err != nil {
		return nil, 0, err
	}
	for rows.Next() {
		err = rows.Scan(&transaction.TransactionId, &transaction.AccountNumber, &transaction.UserId, &transaction.Amount, &transaction.TransferTo, &transaction.CreatedAt, &transaction.UpdatedAt, &transaction.Status, &transaction.Type, &transaction.Comment)
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, transaction)
	}
	rows.Close()
	return transactions, count, nil
}

// Insert adds a new transaction to the database service.
func (d sqlDs) Insert(transaction model.Transaction) error {
	queryString := fmt.Sprintf("INSERT INTO %s", d.table)
	_, err := d.sqlSvc.Exec(queryString+"(user_id, transaction_id, account_number, amount, transfer_to, status, type, comment) VALUES(?,?,?,?,?,?,?,?)", transaction.UserId, transaction.TransactionId, transaction.AccountNumber, transaction.Amount, transaction.TransferTo, transaction.Status, transaction.Type, transaction.Comment)
	if err != nil {
		return err
	}
	return err
}
