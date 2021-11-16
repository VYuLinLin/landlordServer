package controllers

import (
	"encoding/json"
	"fmt"
	"landlord/common"
	"net/http"

	"github.com/astaxie/beego/logs"
)

func LoginOut(w http.ResponseWriter, r *http.Request) {
	var msg = "ok"

	defer func() {
		res := common.Response{
			Msg: msg,
		}
		if msg != "ok" {
			res.Code = 20001
		}
		ret, _ := json.Marshal(res)
		_, err := w.Write(ret)
		if err != nil {
			logs.Error("")
		}
	}()
	username, err := r.Cookie("username")
	if err != nil {
		msg = fmt.Sprintf("request err: %v", err)
		logs.Error(msg)
		return
	}
	userid, err := r.Cookie("userid")
	if err != nil {
		msg = fmt.Sprintf("request err: %v", err)
		logs.Error(msg)
		return
	}
	userData := map[string]interface{}{
		"username": username.Value,
		"id": userid.Value,
	}
	_, err = common.UserInfo(userData)
	if err != nil {
		msg = fmt.Sprintf("logout err: %v", err)
		logs.Debug(msg)
		return
	} else {
		cookie := http.Cookie{Name: "userid", HttpOnly: true, MaxAge: -1}
		http.SetCookie(w, &cookie)
		cookie = http.Cookie{Name: "username", HttpOnly: true, MaxAge: -1}
		http.SetCookie(w, &cookie)
	}
}
