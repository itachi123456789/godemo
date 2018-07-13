package main

import (
	"fmt"
	"godemo/conf"
	"godemo/handler"
	"godemo/module"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Pay Server.")
	engine := gin.New()
	conf.LoadConf()
	handler.Init()

	engine.LoadHTMLGlob("templates/*/*.html")
	engine.Static("/static", "./templates/qr/static")

	// 商品列表
	engine.GET("/", handler.ProductHandler)
	engine.Any("/error", handler.ErrorHandler)

	engine.GET("/qr", handler.QrHandler)

	engine.POST("/pay", handler.PayHandler)

	engine.POST("/order/query", handler.QueryHandler)

	/* -------------- */

	engine.GET("/order", module.OrderHandler)

	// engine.POST("/order/check", module.OrderCheckHandler)
	// engine.POST("/test", module.TestHandler)

	engine.POST("/notify", module.NotifyHandler)

	// http.HandleFunc("/redirect", module.RedirectHandler)

	err := engine.Run(conf.AppConfig.ListenAddr)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
