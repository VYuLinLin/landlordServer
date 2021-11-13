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

func Login(w http.ResponseWriter, r *http.Request) {
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
			logs.Error("")
		}
	}()
	email := r.FormValue("email")
	if len(email) == 0 {
		email = r.PostFormValue("email")
		if email == "" {
			data = "user request Login - err: email is empty"
			logs.Error(data)
			return
		}
	}
	password := r.FormValue("password")
	if len(password) == 0 {
		password = r.PostFormValue("password")
		if password == "" {
			data = "user request Login - err: password is empty"
			logs.Error(data)
			return
		}
	}
	md5Password := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	var account = common.Account{}
	//err := common.GameConfInfo.MysqlConf.Pool.Get(&account, "select * from account where username=? and password", email,md5Password)
	row := common.GameConfInfo.Db.QueryRow("SELECT * FROM `account` WHERE username=? AND password=?", email, md5Password)
	if row != nil {
		err := row.Scan(&account.Id, &account.Email, &account.Username, &account.Password, &account.Coin, &account.CreatedDate, &account.UpdateDate)
		if err != nil {
			data = fmt.Sprintf("user [%v] scan err: %v", email, err)
			logs.Debug(data)
			return
		}
		if account.Id != 0 {
			now := time.Now().Format("2006-01-02 15:04:05")
			_, err := common.GameConfInfo.Db.Exec("UPDATE account SET updated_date = ? WHERE id = ?", now, account.Id)
			if err != nil {
				data = fmt.Sprintf("user [%v] update err: %v", email, err)
				logs.Error(data)
				return
			}
		}

		cookie := http.Cookie{Name: "userid", Value: strconv.Itoa(int(account.Id)), HttpOnly: true, MaxAge: 86400}
		http.SetCookie(w, &cookie)
		cookie = http.Cookie{Name: "username", Value: account.Username, HttpOnly: true, MaxAge: 86400}
		http.SetCookie(w, &cookie)

		data = account
	} else {
		data = fmt.Sprintf("user [%v] request err: user does not exist", email)
		logs.Debug(data)
	}
}
