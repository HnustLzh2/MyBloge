package router

import (
	controller "MyBloge/controllers"
	"MyBloge/docs"
	"MyBloge/tokens"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"os"
	"time"
)

func SetupRouter() *gin.Engine {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	r := gin.Default()
	var SECRET_KEY = os.Getenv("SESSION_KEY")
	if SECRET_KEY == "" {
		SECRET_KEY = "default key"
	}
	store := cookie.NewStore([]byte(SECRET_KEY))
	r.Use(sessions.Sessions("UserSession", store))

	r.Use(cors.New(cors.Config{
		// 允许所有来源
		AllowAllOrigins: true,
		// 允许的 HTTP 方法
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		// 允许的请求头
		AllowHeaders: []string{"Origin", "Content-Type", "Content-Length", "Authorization", "RefreshToken"},
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
	r.GET("/getCategory", controller.GetCategory)
	r.GET("/getCategoryArticle/:category", controller.GetCategoryArticle)
	//GET /articles?text=keyword&page=1&size=10
	r.GET("/searchArticle", controller.SearchArticle)
	article := r.Group("/article")
	article.Use(tokens.Authorization())
	{
		article.POST("/favoriteArticle", controller.FavoriteArticle)
		article.GET("/likeArticle/:articleId", controller.LikeArticle)
		article.POST("/addArticle", controller.AddArticle)
		article.POST("/addComments", controller.AddComment)
		article.POST("/repliedComment", controller.RepliedComment)
		article.POST("/likeComment", controller.LiKeComment)
		article.DELETE("/deleteArticle", controller.DeleteArticle)
		article.DELETE("/deleteComment/:id", controller.DeleteComment)
		article.PUT("/modifyArticle", controller.ModifyArticle)
	}
	folder := r.Group("/folder")
	folder.Use(tokens.Authorization())
	{
		folder.GET("/getArticleFromFolder/:folderId", controller.GetFolderArticles)
		folder.GET("/getMyFolders/:id", controller.GetAllFolders)
		folder.POST("/createFolder", controller.CreateCustomizeFolder)
		folder.POST("/modifyFolder", controller.ModifyCustomizeFolder) //传递的是当前的上下文
		folder.DELETE("/deleteFolder/:folderId", controller.DeleteFolder)
	}
	r.POST("/checkToken", controller.CheckTokenValid)
	r.POST("/refreshToken", controller.RefreshToken)
	// 聊天室相关路由
	chat := r.Group("/chat")
	chat.Use(tokens.Authorization()) // 使用授权中间件
	{
		chat.POST("/create-room", controller.CreateChatRoom)    // 创建聊天室
		chat.GET("/rooms", controller.GetChatRooms)             // 获取所有聊天室
		chat.GET("/room/:roomId", controller.GetChatRoom)       // 获取单个聊天室信息
		chat.GET("/joinRoom", controller.JoinChatRoom)          // 加入聊天室
		chat.GET("/leave-room", controller.LeaveOutCharRoom)    // 真的离开这个聊天室
		chat.DELETE("/room/:roomId", controller.DeleteChatRoom) // 删除聊天室
	}
	// 消息相关路由
	message := r.Group("/message")
	message.Use(tokens.Authorization()) // 使用授权中间件
	{
		message.POST("/send", controller.SendMessage)                 // 发送消息
		message.GET("/history/:roomId", controller.GetMessageHistory) // 获取聊天历史
	}
	// 私聊相关路由
	privateChat := r.Group("/private-chat")
	privateChat.Use(tokens.Authorization()) // 使用授权中间件
	{
		privateChat.POST("/create_private_room", controller.CreatePrivateRoom) //创建私聊房间
		privateChat.GET("/start", controller.StartPrivateChat)                 // 开始私聊
		privateChat.GET("/list", controller.GetPrivateChats)                   // 获取私聊列表
		privateChat.GET("/history/:roomId", controller.GetPrivateChatHistory)  // 获取私聊历史
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
