package main

import (
	"fmt"
	"godemo/conf"
	"godemo/module"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Pay Server.")
	engine := gin.New()
	conf.LoadConf()
	module.Init()

	engine.LoadHTMLGlob("templates/*.html")

	engine.GET("/", module.IndexHandler)
	engine.GET("/order", module.OrderHandler)

	engine.POST("/order/check", module.OrderCheckHandler)
	engine.POST("/test", module.TestHandler)

	engine.Any("/pay", module.PayHandler)
	engine.POST("/notify", module.NotifyHandler)

	http.HandleFunc("/redirect", module.RedirectHandler)

	err := engine.Run(conf.AppConfig.ListenAddr)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
