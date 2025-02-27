package db

import (
	"MyBloge/global"
	"MyBloge/model"
	"MyBloge/utils"
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

func CreateUser(registerRequest utils.RegisterRequest) error {
	var sqlDb = global.SqlDb
	var newUser model.User
	now := time.Now()
	newUser.Name = registerRequest.Username
	newUser.Email = registerRequest.Email
	newUser.Password = registerRequest.Password
	newUser.Authorization = registerRequest.Authorization
	newUser.UserId = uuid.NewSHA1(uuid.NameSpaceDNS, []byte(now.String())).String()
	folder, err := CreateFavoritesFolder(newUser.UserId)
	if err != nil {
		return err
	}
	newUser.FavoritesFolderId = folder.FolderId
	if err := sqlDb.AutoMigrate(&newUser); err != nil {
		return err
	}
	if err := sqlDb.Create(&newUser).Error; err != nil {
		return err
	}
	return nil
}

func CreateFavoritesFolder(id string) (model.FavoritesFolder, error) {
	var sqlDb = global.SqlDb
	var folder model.FavoritesFolder
	folder.UserId = id
	folder.ArticleCollection = []model.BloggerArticle{}
	now := time.Now()
	folder.FolderId = uuid.NewSHA1(uuid.NameSpaceDNS, []byte(now.String())).String()
	if err := sqlDb.AutoMigrate(&folder); err != nil {
		return folder, err
	}
	if err := sqlDb.Create(&folder).Error; err != nil {
		return folder, err
	}
	return folder, nil
}
func CreateArticle(request utils.AddArticleRequest, email string) error {
	var sqlDb = global.SqlDb
	var article = model.BloggerArticle{}
	now := time.Now()
	article.ArticleId = uuid.NewSHA1(uuid.NameSpaceDNS, []byte(now.String())).String()
	article.Title = request.Title
	article.Content = request.Content
	article.Preview = request.Preview
	article.Category = request.Category
	author, err := FindUserByEmail(email)
	if err != nil {
		return err
	}
	article.AuthorId = author.UserId
	if err := sqlDb.AutoMigrate(&article); err != nil {
		return err
	}
	if err := sqlDb.Create(&article).Error; err != nil {
		return err
	}
	return nil
}
func GetAllArticleDB() ([]model.BloggerArticle, error) {
	var sqlDb = global.SqlDb
	var articles []model.BloggerArticle
	if err := sqlDb.Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}
func FindUserByUUID(id string) (model.User, error) {
	var sqlDb = global.SqlDb
	var user model.User
	if err := sqlDb.Where("user_id = ? ", id).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}
func FindArticleByID(id string) (model.BloggerArticle, error) {
	var sqlDb = global.SqlDb
	var article = model.BloggerArticle{}
	if err := sqlDb.Where("article_id = ?", id).First(&article).Error; err != nil {
		return model.BloggerArticle{}, err
	}
	return article, nil
}
func FindUserByEmail(email string) (model.User, error) {
	var sqlDb = global.SqlDb
	var user model.User
	if err := sqlDb.Where("email = ?", email).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func UpdateArticle(article model.BloggerArticle) error {
	var sqlDb = global.SqlDb
	if err := sqlDb.Model(&model.BloggerArticle{}).Where("article_id = ?", article.ArticleId).Updates(article).Error; err != nil {
		return err
	}
	articles, err := GetAllArticleDB()
	if err != nil {
		return err
	}
	articlesJson, err := json.Marshal(articles)
	if err != nil {
		return err
	}
	//更新redis
	if err := SetArticleCache(articlesJson); err != nil {
		return err
	}
	return nil
}

// FavoriteArticleDB 通过使用 GORM 的 Association 方法，你可以正确管理多对多关系的添加和更新操作。确保中间表的配置正确，并使用合适的 GORM 方法来操作多对多关系
func FavoriteArticleDB(articleId string, userID string) error {
	var sqlDb = global.SqlDb
	articleCollection, err := FindCollectionByID(userID)
	article, err := FindArticleByID(articleId)
	if err != nil {
		return err
	}
	// 使用 Association 方法添加 Article 到 ArticleCollection
	if err := sqlDb.Model(&articleCollection).Association("ArticleCollection").Append(&article); err != nil {
		return err
	}
	return nil
}

func FindCollectionByID(userId string) (model.FavoritesFolder, error) {
	var sqlDb = global.SqlDb
	var favoritesFolder model.FavoritesFolder
	if err := sqlDb.Where("user_id = ?", userId).First(&favoritesFolder).Error; err != nil {
		return model.FavoritesFolder{}, err //判断是不是空的
	}
	return favoritesFolder, nil
}

func AddCommentDB(userid string, newComment utils.AddCommentRequest) error {
	var sqlDb = global.SqlDb
	comment := model.Comment{}
	now := time.Now()
	comment.CommentId = uuid.NewSHA1(uuid.NameSpaceDNS, []byte(now.String())).String()
	comment.SendUserId = userid
	comment.Content = newComment.Content
	comment.UserName = newComment.UserName
	comment.UserAvatar = newComment.UserAvatar
	comment.PublishTime = time.Now()
	comment.LikeUserId = []string{}
	comment.RepliedComments = []model.Comment{}
	if err := sqlDb.AutoMigrate(&comment); err != nil {
		return err
	}
	if err := sqlDb.Create(&comment).Error; err != nil {
		return err
	}
	return nil
}

func DeleteArticle(id string) error {
	var sqlDb = global.SqlDb
	if err := sqlDb.Where("article_id = ?", id).Delete(&model.BloggerArticle{}).Error; err != nil {
		return err
	}
	return nil
}
func UpdateUser(user model.User) error {
	var sqlDb = global.SqlDb
	if err := sqlDb.Model(&model.User{}).Where("user_id = ?", user.UserId).Updates(user).Error; err != nil {
		return err
	}
	return nil
}

func GetArticleFromFolder(userID string) (model.FavoritesFolder, error) {
	var sqlDb = global.SqlDb
	var folder model.FavoritesFolder
	if err := sqlDb.Where("user_id = ?", userID).First(&folder).Error; err != nil {
		return folder, err
	}
	return folder, nil
}
