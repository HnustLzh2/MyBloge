package router

import (
	controller "MyBloge/controllers"
	"MyBloge/docs"
	"MyBloge/tokens"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"os"
	"time"
)

func SetupRouter() *gin.Engine {

	r := gin.Default()
	var SECRET_KEY = os.Getenv("SECRET_KEY")
	store := cookie.NewStore([]byte(SECRET_KEY))
	r.Use(sessions.Sessions("UserSession", store))

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
		auth.GET("/logout", controller.Logout)
	}
	r.GET("/getArticle/:id", controller.GetArticleById)
	r.GET("/getAllArticle", controller.GetAllArticle)
	r.GET("/getComment/:id", controller.GetCommentById)
	//GET /articles?text=keyword&page=1&size=10
	r.GET("/searchArticle", controller.SearchArticle)
	r.POST("/likeArticle", controller.LikeArticle)
	article := r.Group("/article")
	article.Use(tokens.Authorization())
	{
		article.POST("/getFavoriteArticle", controller.GetArticleFromFolder)
		article.POST("/addArticle", controller.AddArticle)
		article.POST("/favoriteArticle", controller.FavoriteArticle)

		article.POST("/addComments", controller.AddComment)
		article.POST("/repliedComment", controller.RepliedComment)
		article.POST("/likeComment", controller.LiKeComment)

		article.DELETE("/deleteArticle", controller.DeleteArticle)
		article.PUT("/modifyArticle", controller.ModifyArticle)
	}
	registerSwagger(r)
	return r
}
func registerSwagger(r gin.IRouter) {
	// API文档访问地址: http://host/swagger/index.html
	// 注解定义可参考 https://github.com/swaggo/swag#declarative-comments-format
	// 样例 https://github.com/swaggo/swag/blob/master/example/basic/api/api.go
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Title = "个人管理后台接口"
	docs.SwaggerInfo.Description = "实现一个管理个人博客系统的后端API服务"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:3001"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
