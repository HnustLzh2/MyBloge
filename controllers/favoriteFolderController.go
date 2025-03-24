package controllers

import (
	"MyBloge/db"
	"MyBloge/model"
	"MyBloge/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CreateCustomizeFolder 创建自定义的收藏夹 post
func CreateCustomizeFolder(context *gin.Context) {
	var request = utils.CreateCustomizeFolderRequest{}
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var folder = model.FavoritesFolder{}
	folder, err := db.CreateCustomizeFolderDB(request.FolderName, request.UserId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": folder})
	return
}

// ModifyCustomizeFolder 编辑收藏夹 post
func ModifyCustomizeFolder(context *gin.Context) {
	var request = utils.ModifyCustomizeFolderRequest{}
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.ModifyCustomizeFolder(request.FolderId, request.NewName); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": "Success"})
}

func GetAllFolders(context *gin.Context) {
	var folders []model.FavoritesFolder
	userId := context.Param("id")
	folders, err := db.GetAllFoldersDb(userId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": folders})
}
func GetFolderArticles(context *gin.Context) {
	var articles []model.BloggerArticle
	folderId := context.Param("folderId")
	articles, err := db.GetArticleFromFolder(folderId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": articles})
	return
}

func DeleteFolder(context *gin.Context) {
	var folderId = context.Param("folderId")
	if err := db.DeleteFolderDB(folderId); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"data": "Success"})
	return
}
