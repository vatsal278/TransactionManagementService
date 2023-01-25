package datasource

import (
	"database/sql"
	"fmt"
	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/model"
	"log"
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
		case model.ColumnUpdate:
			a := v.(model.ColumnUpdate)
			f = append(f, fmt.Sprintf("%s = %+v", k, a.UpdateSet))
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

func (d sqlDs) Get(filter map[string]interface{}) ([]model.Account, error) {
	//order the queries based on email address
	var user model.Account
	var users []model.Account
	q := fmt.Sprintf("SELECT user_id, account_number, income, spends, created_on, updated_on, active_services, inactive_services FROM %s", d.table)
	whereQuery := queryFromMap(filter, " AND ")
	if whereQuery != "" {
		q += " WHERE " + whereQuery
	}
	q += " ORDER BY account_number;"
	rows, err := d.sqlSvc.Query(q)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err = rows.Scan(&user.Id, &user.AccountNumber, &user.Income, &user.Spends, &user.CreatedOn, &user.UpdatedOn, &user.ActiveServices, &user.InactiveServices)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (d sqlDs) Insert(user model.Account) error {
	queryString := fmt.Sprintf("INSERT INTO %s", d.table)
	log.Print("(user_id, active_services, inactive_services) VALUES(?,?,?)", user.Id, user.ActiveServices, user.InactiveServices)
	_, err := d.sqlSvc.Exec(queryString+"(user_id, active_services, inactive_services) VALUES(?,?,?)", user.Id, user.ActiveServices, user.InactiveServices)
	if err != nil {
		return err
	}
	return err
}

func (d sqlDs) Update(filterSet map[string]interface{}, filterWhere map[string]interface{}) error {
	queryString := fmt.Sprintf("UPDATE %s ", d.table)

	setQuery := queryFromMap(filterSet, " , ")
	if setQuery != "" {
		queryString += " SET " + setQuery
	}
	whereQuery := queryFromMap(filterWhere, " AND ")
	if whereQuery != "" {
		queryString += " WHERE " + whereQuery
	}
	queryString += " ;"
	log.Print(queryString)
	_, err := d.sqlSvc.Exec(queryString)
	if err != nil {
		return err
	}
	return nil
}
