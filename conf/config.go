package conf

import (
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"landlord/common"
	"os"
)

var (
	GameConf = &common.GameConfInfo
)

func InitConf() (err error) {
	environment := os.Getenv("ENV")
	if environment != "dev" && environment != "testing" && environment != "product" {
		environment = "product"
	}
	logs.Info("the running environment is : %s", environment)
	conf, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		logs.Error("new conf failed ,err : %v", err)
		return
	}

	environment += "::"
	GameConf.HttpPort, err = conf.Int(environment + "http_port")
	if err != nil {
		logs.Error("config http_port failed,err: %v", err)
		return
	}

	logs.Debug("read conf succ , http port : %v", GameConf.HttpPort)

	//todo 日志配置
	GameConf.LogPath = conf.String(environment + "log_path")
	if len(GameConf.LogPath) == 0 {
		GameConf.LogPath = "./logs/game.log"
	}

	logs.Debug("read conf succ , LogPath :  %v", GameConf.LogPath)
	GameConf.LogLevel = conf.String(environment + "log_level")
	if len(GameConf.LogLevel) == 0 {
		GameConf.LogLevel = "debug"
	}
	logs.Debug("read conf succ , LogLevel :  %v", GameConf.LogLevel)

	//todo sqlite配置
	GameConf.DbPath = conf.String(environment + "db_path")
	if len(GameConf.DbPath) == 0 {
		GameConf.DbPath = "./db/landlord.db"
	}
	logs.Debug("read conf succ , DbPath :  %v", GameConf.DbPath)
	return
}
