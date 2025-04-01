package main

import (
	"MyBloge/config"
	controller "MyBloge/controllers"
	"MyBloge/db"
	_ "MyBloge/docs"
	"MyBloge/router"
	"github.com/gin-gonic/gin"
)

// 在同一个包下，创建一个测试文件 math_test.go 测试文件的名称必须以 _test.go 结尾，这样 Go 才能识别它为测试文件。
// 这个项目大致完成了，我要开始更新前端了
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
