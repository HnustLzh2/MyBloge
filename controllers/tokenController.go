package controllers

import (
	"MyBloge/tokens"
	"MyBloge/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CheckTokenValid(context *gin.Context) {
	var request utils.TokenRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := tokens.VerifyToken(request.AuthToken)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	_, err = tokens.VerifyToken(request.RefreshToken)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"token": "Right Token"})
	return
}
func RefreshToken(context *gin.Context) {
	var request utils.TokenRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := tokens.VerifyToken(request.RefreshToken)
	if err != nil {
		context.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	tokenString, RefreshTokenString, err := tokens.RefreshToken(request.RefreshToken)
	if err != nil {
		context.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	//给session和cookies设置
	context.SetCookie("authToken", tokenString, 60*60*1, "/", "", false, true)
	context.SetCookie("refreshToken", RefreshTokenString, 60*60*24*7, "/", "", false, true)
	context.JSON(http.StatusOK, gin.H{"AuthToken": tokenString, "RefreshToken": RefreshTokenString})
	return
}
