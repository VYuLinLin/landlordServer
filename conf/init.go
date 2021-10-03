package conf

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	_ "github.com/mattn/go-sqlite3"
)

func conversionLogLevel(logLevel string) int {
	switch logLevel {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}
	return logs.LevelDebug
}

func initLogger() (err error) {
	config := make(map[string]interface{})
	config["filename"] = GameConf.LogPath
	config["level"] = conversionLogLevel(GameConf.LogLevel)

	configStr, err := json.Marshal(config)
	if err != nil {
		fmt.Println("marsha1 failed,err", err)
		return
	}
	err = logs.SetLogger(logs.AdapterFile, string(configStr))
	if err != nil {
		logs.Error("init logger failed:%v", err)
	}
	return
}

func initSqlite() (err error) {
	GameConf.Db, err = sql.Open("sqlite3", GameConf.DbPath)
	if err != nil {
		logs.Error("initSqlite err : %v", err)
		return
	}
	var stmt *sql.Stmt
	stmt, err = GameConf.Db.Prepare(`CREATE TABLE IF NOT EXISTS "account" ("id" INTEGER NOT NULL,"email" text(32),"username" TEXT(16),"password" TEXT(32),"coin" integer,"created_date" TEXT(32),"updated_date" TEXT(32),PRIMARY KEY ("id"))`)
	if err != nil {
		logs.Error("initSqlite err : %v", err)
		return
	}
	_, err = stmt.Exec()
	if err != nil {
		logs.Error("create table err:", err)
		return
	}
	return
}

func InitSec() (err error) {
	err = initLogger()
	if err != nil {
		logs.Error("init logger failed,err:%v", err)
		return
	}
	err = initSqlite()
	if err != nil {
		logs.Error("init sqlite failed,err:%v", err)
		return
	}
	logs.Info("init sec success")
	return
}
