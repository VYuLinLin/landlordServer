package common

import (
	"database/sql"
	"fmt"
	"strings"
)

// UserInfo 获取用户信息
func UserInfo(userMap map[string]interface{}) (account Account, err error) {
	var row *sql.Row
	defer func() {
		if row == nil {
			err = fmt.Errorf("err: user does not exist")
		}
	}()
	var where []string
	var values []interface{}
	query := "SELECT * FROM `account` WHERE "
	for s, i := range userMap {
		where = append(where, fmt.Sprintf("%s = ?", s))
		values = append(values, i)
	}
	query += strings.Join(where, " AND ")
	row = GameConfInfo.Db.QueryRow(query, values...)
	err = row.Scan(&account.Id, &account.Email, &account.Username, &account.Password, &account.Coin, &account.CreatedDate, &account.UpdateDate)
	return
}
