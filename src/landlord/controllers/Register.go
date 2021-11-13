package controllers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"landlord/common"
	"net/http"
	"strconv"
	"time"

	"github.com/astaxie/beego/logs"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var data interface{}
	res := make(map[string]interface{})

	defer func() {
		// w.Header().Set("Access-Control-Allow-Origin", "*")
		res["data"] = data
		if _, ok := data.(string); ok {
			res["status"] = false
		} else {
			res["status"] = true
		}
		ret, _ := json.Marshal(res)
		_, err := w.Write(ret)
		if err != nil {
			logs.Error("user request Register - err : %v", err)
		}
	}()
	username := r.FormValue("username")
	if len(username) == 0 {
		username = r.PostFormValue("username")
		if username == "" {
			data = "register err: username is empty"
			logs.Error(data)
			return
		}
	}
	password := r.FormValue("password")
	if len(password) == 0 {
		password = r.PostFormValue("password")
		if password == "" {
			data = "register err: password is empty"
			logs.Error(data)
			return
		}
	}
	logs.Debug("user request register : username=%v, password=%v ", username, password)

	var account = common.Account{}
	row := common.GameConfInfo.Db.QueryRow("select * from account where username=?", username)
	err := row.Scan(&account.Id, &account.Email, &account.Username, &account.Password, &account.Coin, &account.CreatedDate, &account.UpdateDate)
	if err != nil {
		now := time.Now().Format("2006-01-02 15:04:05")
		md5Password := fmt.Sprintf("%x", md5.Sum([]byte(password)))
		stmt, err := common.GameConfInfo.Db.Prepare(`insert into account(email,username,password,coin,created_date,updated_date) values(?,?,?,?,?,?) `)
		if err != nil {
			data = fmt.Sprintf("insert into account [%v] err : %v", username, err)
			logs.Error(data)
			return
		}
		defer stmt.Close()
		result, err := stmt.Exec(username, username, md5Password, 10000, now, now)
		if err != nil {
			data = fmt.Sprintf("insert new account [%v] err : %v", username, err)
			logs.Error(data)
			return
		}
		lastInsertId, err := result.LastInsertId()
		if err != nil {
			data = fmt.Sprintf("insert new account [%v] get last insert id err : %v", username, err)
			logs.Error(data)
			return
		}

		cookie := http.Cookie{Name: "userid", Value: strconv.Itoa(int(lastInsertId)), HttpOnly: true, MaxAge: 86400}
		http.SetCookie(w, &cookie)
		cookie = http.Cookie{Name: "username", Value: username, HttpOnly: true, MaxAge: 86400}
		http.SetCookie(w, &cookie)

		data = map[string]interface{}{
			"uid":          lastInsertId,
			"username":     username,
			"coin":         1000,
			"created_date": now,
			"updated_date": now,
		}
	} else {
		data = fmt.Sprintf("user [%v] request register err: user already exists", username)
		logs.Debug(data)
	}
}
