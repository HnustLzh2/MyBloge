package controllers

import (
	"MyBloge/db"
	"MyBloge/model"
	"MyBloge/utils"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"net/http"
	"strconv"
)

func GetArticleById(context *gin.Context) {
	var articleId = context.Param("id")
	var article = model.BloggerArticle{}
	article, err := db.FindArticleByID(articleId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"article": article})
}

func AddArticle(context *gin.Context) {
	var request utils.AddArticleRequest
	email, _ := context.Get("email")
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	if err := db.FavoriteArticleDB(request.ArticleId, request.UserId); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": "Add successfully"})
}

func LikeArticle(context *gin.Context) {
	id := context.Param("articleId")
	article, err := db.FindArticleByID(id)
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
		context.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	if err := db.AddCommentDB(request.SendUserId, request); err != nil {
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
	if err := db.DeleteArticle(request.ArticleId); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	//这里删除之后要删除redis缓存中的数据
	if err := db.DeleteArticleCache(); err != nil {
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
	newArticle, err := db.FindArticleByID(request.ID)
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
	context.JSON(http.StatusOK, gin.H{"success": "Update successfully"})
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
		context.JSON(http.StatusOK, gin.H{"articles": articles})
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
		context.JSON(http.StatusOK, gin.H{"articles": articles})
	}
}

func GetArticleFromFolder(context *gin.Context) {
	//拿到userId
	var request utils.UserIdRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	folder, err := db.GetArticleFromFolder(request.UserId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": folder})
}

func RepliedComment(context *gin.Context) {
	var request utils.RepliedCommentRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.RepliedCommentDb(request); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := db.DeleteCommentCache(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": "Replied successfully"})
}

func LiKeComment(context *gin.Context) {
	var request utils.LikeCommentRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.LikeCommentDB(request.CommentId, request.UserId); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//更新redis
	if err := db.DeleteCommentCache(); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": "Like successfully"})
}

func GetCommentById(context *gin.Context) {
	var articleId = context.Param("id")
	var comment []model.Comment
	commentsCache, err := db.GetCommentCache()
	if errors.Is(err, redis.Nil) {
		if err := db.GetCommentDB(articleId, &comment); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		commentByte, err := json.Marshal(comment)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := db.SetCommentCache(commentByte); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		context.JSON(http.StatusOK, gin.H{"success": comment})
	} else if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		//[]byte()括号里要是string类型，就是把string变成字节类型，存进redis缓存中
		if err := json.Unmarshal([]byte(commentsCache), &comment); err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		context.JSON(http.StatusOK, gin.H{"success": comment})
	}
	context.JSON(http.StatusOK, gin.H{"success": comment})
}

func SearchArticle(c *gin.Context) {
	text := c.Query("text")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	articles, total, err := db.SearchArticle(text, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}
	c.JSON(http.StatusOK, gin.H{"success": articles, "total": total, "size": size, "page": page})
}
