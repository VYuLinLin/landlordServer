package router

import (
	"landlord/controllers"
	"landlord/service"
	"log"
	"net/http"
)

type HandleFunc func(http.ResponseWriter, *http.Request)

func init() {
	http.HandleFunc("/", logPanics(controllers.Index))
	http.HandleFunc("/login", logPanics(controllers.Login))
	http.HandleFunc("/loginOut", logPanics(controllers.LoginOut))
	http.HandleFunc("/reg", logPanics(controllers.Register))

	http.HandleFunc("/ws", service.ServeWs)

	// 设置静态目录
	static := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", static))
}

func logPanics(f HandleFunc) HandleFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if x := recover(); x != nil {
				log.Printf("[%v] caught panic: %v", req.RemoteAddr, x)
				// 给页面一个错误信息, 如下示例返回一个500
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		f(w, req)
	}
}
