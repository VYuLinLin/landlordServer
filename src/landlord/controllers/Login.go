package controllers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
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
	var msg string = "ok"

	defer func() {
		res := common.Response{
			0,
			msg,
			data,
		}
		if data == nil {
			res.Code = 20001
		}
		ret, _ := json.Marshal(res)
		_, err := w.Write(ret)
		if err != nil {
			logs.Error("")
		}
	}()

	defer r.Body.Close()
	reqByte, _ := ioutil.ReadAll(r.Body)
	reqData := &reqData{}
	json.Unmarshal(reqByte, reqData)

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
	userInfo := common.Account{}
	row := common.GameConfInfo.Db.QueryRow("SELECT * FROM `account` WHERE username=? AND password=?", account, md5Password)
	if row != nil {
		err := row.Scan(&userInfo.Id, &userInfo.Email, &userInfo.Username, &userInfo.Password, &userInfo.Coin, &userInfo.CreatedDate, &userInfo.UpdateDate)
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

		cookie := http.Cookie{Name: "userid", Value: strconv.Itoa(int(userInfo.Id)), HttpOnly: true, MaxAge: 86400}
		http.SetCookie(w, &cookie)
		cookie = http.Cookie{Name: "username", Value: userInfo.Username, HttpOnly: true, MaxAge: 86400}
		http.SetCookie(w, &cookie)

		data = userInfo
	} else {
		msg = fmt.Sprintf("user [%v] request err: user does not exist", account)
		logs.Debug(msg)
	}
}
