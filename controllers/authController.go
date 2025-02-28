package controllers

import (
	"MyBloge/db"
	"MyBloge/tokens"
	"MyBloge/utils"
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
	if err := tokens.UpdateToken(tokenString, refreshToken, &user); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := db.UpdateUser(user); err != nil {
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
