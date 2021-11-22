package router

import (
	"landlord/controllers"
	"landlord/service"
	"log"
	"net/http"
)

type HandleFunc func(http.ResponseWriter, *http.Request)

func init() {
	http.HandleFunc("/login", logPanics(controllers.Login))
	http.HandleFunc("/logout", logPanics(controllers.LoginOut))
	http.HandleFunc("/register", logPanics(controllers.Register))

	http.HandleFunc("/ws/", logPanics(service.ServeWs))
	http.HandleFunc("/", logPanics(controllers.Index))

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
		// 跨域请求时，允许头部携带cookie，设置后Access-Control-Allow-Origin值不能是“*”
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", "http://172.21.165.80:7456")
		//w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
		} else {
			f(w, req)
		}
	}
}
