package controllers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"landlord/common"
	"net/http"
	"strconv"
	"time"

	"github.com/astaxie/beego/logs"
)

type reqData struct {
	Account  string
	Password string
}

func Login(w http.ResponseWriter, r *http.Request) {
	var data interface{}
	var msg = "ok"

	defer func() {
		res := common.Response{
			Msg:  msg,
			Data: data,
		}
		if data == nil {
			res.Code = 20001
		}
		ret, _ := json.Marshal(res)
		_, err := w.Write(ret)
		if err != nil {
			logs.Error("login - err : %v", err)
		}
	}()

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			
		}
	}(r.Body)
	reqByte, _ := ioutil.ReadAll(r.Body)
	reqData := &reqData{}
	err := json.Unmarshal(reqByte, reqData)
	if err != nil {
		msg = fmt.Sprintf("json unmarshal err : %v", err)
		logs.Error(msg)
		return 
	}

	account := reqData.Account
	if account == "" {
		msg = "user request Login - err: account is empty"
		logs.Error(msg)
		return
	}
	password := reqData.Password
	if password == "" {
		msg = "user request Login - err: password is empty"
		logs.Error(msg)
		return
	}
	md5Password := fmt.Sprintf("%x", md5.Sum([]byte(password)))

	userData := map[string]interface{}{
		"username": account,
		"password": md5Password,
	}
	userInfo, err := common.UserInfo(userData)
	if err != nil {
		msg = fmt.Sprintf("user [%v] scan err: %v", account, err)
		logs.Debug(msg)
		return
	}
	if userInfo.Id != 0 {
		now := time.Now().Format("2006-01-02 15:04:05")
		_, err := common.GameConfInfo.Db.Exec("UPDATE account SET updated_date = ? WHERE id = ?", now, userInfo.Id)
		if err != nil {
			msg = fmt.Sprintf("user [%v] update err: %v", account, err)
			logs.Error(msg)
			return
		}
	}

	cookie := http.Cookie{Name: "userid", Value: strconv.Itoa(userInfo.Id), HttpOnly: true, MaxAge: 86400}
	http.SetCookie(w, &cookie)
	cookie = http.Cookie{Name: "username", Value: userInfo.Username, HttpOnly: true, MaxAge: 86400}
	http.SetCookie(w, &cookie)

	data = userInfo
}
