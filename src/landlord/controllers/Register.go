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

func Register(w http.ResponseWriter, r *http.Request) {
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
			logs.Error("user request Register - err : %v", err)
		}
	}()
	defer r.Body.Close()
	reqByte, _ := ioutil.ReadAll(r.Body)
	reqData := &reqData{}
	json.Unmarshal(reqByte, reqData)

	account := reqData.Account
	if account == "" {
		msg = "register err: account is empty"
		logs.Error(msg)
		return
	}

	password := reqData.Password
	if password == "" {
		msg = "register err: password is empty"
		logs.Error(msg)
		return
	}
	logs.Debug("user request register : account=%v, password=%v ", account, password)

	userInfo := common.Account{}
	row := common.GameConfInfo.Db.QueryRow("select * from account where username=?", account)
	err := row.Scan(&userInfo.Id, &userInfo.Email, &userInfo.Username, &userInfo.Password, &userInfo.Coin, &userInfo.CreatedDate, &userInfo.UpdateDate)
	if err != nil {
		now := time.Now().Format("2006-01-02 15:04:05")
		md5Password := fmt.Sprintf("%x", md5.Sum([]byte(password)))
		stmt, err := common.GameConfInfo.Db.Prepare(`insert into account(email,username,password,coin,created_date,updated_date) values(?,?,?,?,?,?) `)
		if err != nil {
			msg = fmt.Sprintf("insert into account [%v] err : %v", account, err)
			logs.Error(msg)
			return
		}
		defer stmt.Close()
		result, err := stmt.Exec(account, account, md5Password, 10000, now, now)
		if err != nil {
			msg = fmt.Sprintf("insert new account [%v] err : %v", account, err)
			logs.Error(msg)
			return
		}
		lastInsertId, err := result.LastInsertId()
		if err != nil {
			msg = fmt.Sprintf("insert new account [%v] get last insert id err : %v", account, err)
			logs.Error(msg)
			return
		}

		cookie := http.Cookie{Name: "userid", Value: strconv.Itoa(int(lastInsertId)), HttpOnly: true, MaxAge: 86400}
		http.SetCookie(w, &cookie)
		cookie = http.Cookie{Name: "username", Value: account, HttpOnly: true, MaxAge: 86400}
		http.SetCookie(w, &cookie)

		data = map[string]interface{}{
			"id":           lastInsertId,
			"username":     account,
			"coin":         1000,
			"created_date": now,
			"updated_date": now,
		}
	} else {
		msg = "register err: user already exists"
		logs.Debug(msg)
	}
}
