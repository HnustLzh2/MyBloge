package main

import (
	"MyBloge/config"
	"MyBloge/router"
	"github.com/gin-gonic/gin"
)

func main() {
	config.InitConfig()
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
