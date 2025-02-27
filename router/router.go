package router

import (
	controller "MyBloge/controllers"
	"MyBloge/tokens"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		// 允许所有来源
		AllowAllOrigins: true,
		// 允许的 HTTP 方法
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		// 允许的请求头
		AllowHeaders: []string{"Origin", "Content-Type", "Content-Length", "Authorization"},
		// 允许的响应头
		ExposeHeaders: []string{"Content-Length"},
		// 允许携带凭证
		AllowCredentials: true,
		// 预检请求缓存时间
		MaxAge: 12 * time.Hour,
	}))
	auth := r.Group("/auth")
	{
		auth.POST("/login", controller.Login)
		auth.POST("/register", controller.Register)
	}
	r.GET("/getArticle/:id", controller.GetArticleById)
	r.GET("/getAllArticle", controller.GetAllArticle)
	r.GET("/getComment/:id", controller.GetCommentById)
	article := r.Group("/article")
	article.Use(tokens.Authentication())
	{
		article.POST("/getFavoriteArticle", controller.GetArticleFromFolder)
		article.POST("/addArticle", controller.AddArticle)
		article.POST("/favoriteArticle", controller.FavoriteArticle)
		article.POST("/likeArticle", controller.LikeArticle)
		article.POST("/addComments", controller.AddComment)
		article.POST("/repliedComment", controller.RepliedComment)
		article.POST("/likeComment", controller.LiKeComment)

		article.DELETE("/deleteArticle", controller.DeleteArticle)

		article.PUT("/modifyArticle", controller.ModifyArticle)
	}
	return r
}
