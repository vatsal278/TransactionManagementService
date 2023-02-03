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

//docker run --rm --env MYSQL_ROOT_PASSWORD=pass --env MYSQL_DATABASE=accmgmt --publish 9085:3306 --name mysqlDb -d mysql
func NewSql(dbSvc config.DbSvc, tableName string) DataSourceI {
	return &sqlDs{
		sqlSvc: dbSvc.Db,
		table:  tableName,
	}
}

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

func (d sqlDs) HealthCheck() bool {
	err := d.sqlSvc.Ping()
	if err != nil {
		return false
	}
	return true
}

func (d sqlDs) Get(filter map[string]interface{}, limit int, offset int) ([]model.Transaction, int, error) {
	//order the queries based on email address
	var transaction model.Transaction
	var transactions []model.Transaction
	var count int
	//include user_id in model and also query it for internal use but never pass it into response
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
	if limit > 0 || offset >= 0 {
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

func (d sqlDs) Insert(transaction model.Transaction) error {
	queryString := fmt.Sprintf("INSERT INTO %s", d.table)
	_, err := d.sqlSvc.Exec(queryString+"(user_id, transaction_id, account_number, amount, transfer_to, status, type, comment) VALUES(?,?,?,?,?,?,?,?)", transaction.UserId, transaction.TransactionId, transaction.AccountNumber, transaction.Amount, transaction.TransferTo, transaction.Status, transaction.Type, transaction.Comment)
	if err != nil {
		return err
	}
	return err
}
