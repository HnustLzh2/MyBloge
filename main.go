package main

import (
	"MyBloge/config"
	controller "MyBloge/controllers"
	"MyBloge/db"
	"MyBloge/router"
	"github.com/gin-gonic/gin"
)

func main() {
	config.InitConfig()
	db.InitRedisValue()
	db.InitDbOperate()
	controller.InitAuthValue()
	r := router.SetupRouter()
	port := config.AppConfig.App.Port
	if port == "" {
		port = "8080"
	}
	r.Use(gin.Logger())
	err := r.Run(":" + port)
	if err != nil {
		panic(err)
	}
}
