package controllers

import (
	"encoding/json"
	"fmt"
	"landlord/common"
	"net/http"

	"github.com/astaxie/beego/logs"
)

func LoginOut(w http.ResponseWriter, r *http.Request) {
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
	username, err := r.Cookie("username")
	if err != nil {
		data = fmt.Sprintf("request err: %v", err)
		logs.Error(data)
		return
	}
	userid, err := r.Cookie("userid")
	if err != nil {
		data = fmt.Sprintf("request err: %v", err)
		logs.Error(data)
		return
	}
	row := common.GameConfInfo.Db.QueryRow("SELECT * FROM `account` WHERE username=? AND id=?", username.Value, userid.Value)
	var account = common.Account{}
	err = row.Scan(&account.Id, &account.Email, &account.Username, &account.Password, &account.Coin, &account.CreatedDate, &account.UpdateDate)
	if err != nil {
		data = fmt.Sprintf("user err: %v", err)
		logs.Debug(data)
		return
	} else {
		cookie := http.Cookie{Name: "userid", Path: "/", MaxAge: -1}
		http.SetCookie(w, &cookie)
		cookie = http.Cookie{Name: "username", Path: "/", MaxAge: -1}
		http.SetCookie(w, &cookie)
	}

}
