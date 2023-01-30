package datasource

import (
	"database/sql"
	"fmt"
	"github.com/PereRohit/util/log"
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

func (d sqlDs) Get(filter map[string]interface{}, limit int, offset int) ([]model.GetTransaction, int, error) {
	//order the queries based on email address
	var transaction model.GetTransaction
	var transactions []model.GetTransaction
	var count int
	q := fmt.Sprintf("SELECT transaction_id, account_number, amount, transfer_to, created_at, updated_at, status, type, comment FROM %s", d.table)
	whereQuery := queryFromMap(filter, " AND ")
	if whereQuery != "" {
		q += " WHERE " + whereQuery
	}
	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM %s", d.table) + " WHERE " + whereQuery
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
	rowsCount.Close()
	log.Info(count)
	q += fmt.Sprintf(" ORDER BY created_at LIMIT %d OFFSET %d ;", limit, offset)
	rows, err := d.sqlSvc.Query(q)
	if err != nil {
		return nil, 0, err
	}
	for rows.Next() {
		err = rows.Scan(&transaction.TransactionId, &transaction.AccountNumber, &transaction.Amount, &transaction.TransferTo, &transaction.CreatedAt, &transaction.UpdatedAt, &transaction.Status, &transaction.Type, &transaction.Comment)
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, transaction)
	}
	rows.Close()
	return transactions, count, nil
}

func (d sqlDs) Insert(newTransaction model.NewTransaction) error {
	queryString := fmt.Sprintf("INSERT INTO %s", d.table)
	_, err := d.sqlSvc.Exec(queryString+"(user_id, transaction_id, account_number, amount, transfer_to, status, type, comment) VALUES(?,?,?,?,?,?,?,?)", newTransaction.UserId, newTransaction.TransactionId, newTransaction.AccountNumber, newTransaction.Amount, newTransaction.TransferTo, newTransaction.Status, newTransaction.Type, newTransaction.Comment)
	if err != nil {
		return err
	}
	return err
}
