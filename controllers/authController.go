package controllers

import (
	"MyBloge/db"
	"MyBloge/tokens"
	"MyBloge/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var loginLimit int
var loginTimer *time.Timer
var isLoginCooldown bool

func InitAuthValue() {
	loginLimit = 0
	isLoginCooldown = false
	loginTimer = time.NewTimer(0)
}
func reloadAuthValue() {
	loginLimit = 0
	isLoginCooldown = false
	loginTimer.Reset(0)
}

func Login(context *gin.Context) {
	if isLoginCooldown {
		context.JSON(http.StatusForbidden, gin.H{"error": "抱歉你的登入次数过多，请等待3分钟后再来重试"})
		return
	}
	var loginRequest = utils.LoginRequest{}
	if err := context.ShouldBindJSON(&loginRequest); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	loginLimit++
	user, err := db.FindUserByEmail(loginRequest.Email)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	if !utils.CheckOutPassword(user.Password, loginRequest.Password) {
		loginLimit++
		if loginLimit > 3 {
			// 启动冷却期
			isLoginCooldown = true
			loginTimer.Reset(3 * time.Minute) // 重置定时器为3分钟
			go func() {
				<-loginTimer.C    // 等待定时器触发
				reloadAuthValue() // 重置登录限制
			}()
		}
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong password"})
		return
	}
	//登入成功，重置次数
	reloadAuthValue()
	//更新token
	tokenString, refreshToken, err := tokens.GenerateToken(user.Email, user.Name)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tokens.UpdateToken(tokenString, refreshToken, &user)
	if err := db.UpdateUser(user); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//name：Cookie 的名称，这里是 "authToken"。
	//value：Cookie 的值，这里是 tokenString。
	//maxAge：Cookie 的最大年龄（以秒为单位）。如果设置为负值，Cookie 会被标记为会话 Cookie，这意味着它会在浏览器关闭时失效。如果设置为 0，Cookie 会被删除。
	//在你的例子中，60 * 60 * 24 表示 Cookie 的有效期为 24 小时。
	//path：Cookie 的路径，这里是 "/"，表示 Cookie 在整个网站上都有效。
	//domain：Cookie 的域名，这里是空字符串 ""，表示 Cookie 仅在当前域名下有效。
	//secure：是否仅通过 HTTPS 传输 Cookie，这里是 false，表示在 HTTP 和 HTTPS 下都可以传输。
	//httpOnly：是否将 Cookie 设置为 HTTPOnly，这里是 true，表示 Cookie 不可被 JavaScript 访问，从而提高安全性。
	context.SetCookie("authToken", tokenString, 60*60*1, "/", "", false, true)
	context.SetCookie("refreshToken", refreshToken, 60*60*24*7, "/", "", false, true)
	context.Header("RefreshToken", refreshToken)
	context.Header("Authorization", tokenString)
	session := sessions.Default(context)
	session.Set("Authorization", user.Authorization)
	session.Set("UserId", user.UserId)
	if err := session.Save(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"user": user})
}

func Register(context *gin.Context) {
	var registerRequest = utils.RegisterRequest{}
	if err := context.ShouldBindJSON(&registerRequest); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ok := utils.CheckEmailValid(registerRequest.Email)
	if !ok {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Wrong email"})
		return
	}
	_, err := db.FindUserByEmail(registerRequest.Email)
	if err == nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		return
	}
	ok = utils.CheckPasswordValid(registerRequest.Password)
	if !ok {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Valid password",
			"right password": "密码必须有6位以上，而且密码要求至少要大小写字母符号其中三个才行"},
		)
		return
	}
	password, err := utils.EncodePassword(registerRequest.Password)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	registerRequest.Password = password
	if err := db.CreateUser(registerRequest); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"Success": "Successfully Register User", "User": registerRequest})
}

func Logout(context *gin.Context) {
	//退出登入
	tokens.NullifyTokenCookiesAndHeader(context)
	session := sessions.Default(context)
	session.Delete("Authorization")
	session.Delete("UserId")
	session.Clear()
	if err := session.Save(); err != nil {
		return
	}
	context.JSON(http.StatusOK, gin.H{"Success": "Successfully Logout User"})
}
