package controllers

import (
	"MyBloge/db"
	"MyBloge/model"
	"MyBloge/utils"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"net/http"
)

func GetArticleById(context *gin.Context) {
	var articleId = context.Param("id")
	var article = model.BloggerArticle{}
	parseID, err := uuid.Parse(articleId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	article, err = db.FindArticleByID(parseID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	context.JSON(http.StatusOK, gin.H{"article": article})
}

func AddArticle(context *gin.Context) {
	var request utils.AddArticleRequest
	email, _ := context.Get("email")
	if err := db.CreateArticle(request, email.(string)); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"article": request})
}

func FavoriteArticle(context *gin.Context) {
	var request utils.FavoriteArticleRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId := request.UserId
	userParseID, err := uuid.Parse(userId)
	articleId := request.ArticleId
	articleParseID, err := uuid.Parse(articleId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.FavoriteArticleDB(articleParseID, userParseID); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": "Add successfully"})
}

func LikeArticle(context *gin.Context) {
	id := context.Param("id")
	parseID, err := uuid.Parse(id)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	article, err := db.FindArticleByID(parseID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	article.LikesNum++
	err = db.UpdateArticle(article)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": "Like successfully"})
}

func AddComment(context *gin.Context) {
	var request utils.AddCommentRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	parseID, err := uuid.Parse(request.SendUserId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.AddCommentDB(parseID, request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": "Add successfully"})
}

func DeleteArticle(context *gin.Context) {
	var request utils.ArticleIdRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	parseID, err := uuid.Parse(request.ArticleId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.DeleteArticle(parseID); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": "Delete successfully"})
}

func ModifyArticle(context *gin.Context) {
	var request utils.ModifyArticleRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	parseID, err := uuid.Parse(request.ID)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newArticle, err := db.FindArticleByID(parseID)
	newArticle.Preview = request.Preview
	newArticle.Title = request.Title
	newArticle.Content = request.Content
	newArticle.Category = request.Category
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := db.UpdateArticle(newArticle); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func GetAllArticle(context *gin.Context) {
	var articles []model.BloggerArticle
	articleString, err := db.GetArticleFromRedis()
	if errors.Is(err, redis.Nil) {
		articles, err := db.GetAllArticleDB()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		articleJson, err := json.Marshal(articles)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := db.SetArticleCache(articleJson); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		context.JSON(http.StatusOK, gin.H{"article": articles})
		return
	} else if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		err = json.Unmarshal([]byte(articleString), &articles)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		context.JSON(http.StatusOK, gin.H{"article": articles})
	}
}
