package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"landlord/common"
	"landlord/service"
	"net/http"
	"strconv"
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

	params := r.URL.Query()
	userId, err := strconv.Atoi(params.Get("id"))
	userName := params.Get("name")

	if err != nil {
		msg = fmt.Sprintf("logout: id conversion err: %v", err)
		logs.Error(msg)
		return
	}
	userData := map[string]interface{}{
		"username": userName,
		"id": userId,
	}
	_, err = common.UserInfo(userData)
	if err != nil {
		msg = fmt.Sprintf("logout err: %v", err)
		logs.Debug(msg)
		return
	}
	service.DeleteClient(service.UserId(userId))
}
