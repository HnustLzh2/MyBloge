package controllers

import (
	"MyBloge/db"
	"MyBloge/tokens"
	"MyBloge/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Login(context *gin.Context) {
	var loginRequest = utils.LoginRequest{}
	if err := context.ShouldBindJSON(&loginRequest); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := db.FindUserByEmail(loginRequest.Email)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	if !utils.CheckOutPassword(user.Password, loginRequest.Password) {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong password"})
		return
	}
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
