package db

import (
	"MyBloge/global"
	"MyBloge/model"
	"MyBloge/utils"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

var sqlDb *gorm.DB

func InitDbOperate() {
	sqlDb = global.SqlDb
}
func CreateUser(registerRequest utils.RegisterRequest) error {
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
	var folder model.FavoritesFolder
	folder.UserId = id
	folder.ArticleCollection = []model.BloggerArticle{}
	folder.FolderName = "默认收藏夹"
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
	if err = DeleteArticleCache(); err != nil {
		return err
	}
	return nil
}
func GetAllArticleDB() ([]model.BloggerArticle, error) {
	var articles []model.BloggerArticle
	if err := sqlDb.Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}
func FindUserByUUID(id string) (model.User, error) {
	var user model.User
	if err := sqlDb.Where("user_id = ? ", id).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}
func FindArticleByID(id string) (model.BloggerArticle, error) {
	var article = model.BloggerArticle{}
	if err := sqlDb.Where("article_id = ?", id).First(&article).Error; err != nil {
		return model.BloggerArticle{}, err
	}
	return article, nil
}
func FindUserByEmail(email string) (model.User, error) {
	var user model.User
	if err := sqlDb.Where("email = ?", email).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func UpdateArticle(article model.BloggerArticle) error {
	if err := sqlDb.Model(&model.BloggerArticle{}).Where("article_id = ?", article.ArticleId).Updates(article).Error; err != nil {
		return err
	}
	//更新redis  **
	if err := DeleteArticleCache(); err != nil {
		return err
	}
	return nil
}

// FavoriteArticleDB 通过使用 GORM 的 Association 方法，你可以正确管理多对多关系的添加和更新操作。确保中间表的配置正确，并使用合适的 GORM 方法来操作多对多关系
func FavoriteArticleDB(articleId string, userID string) error {
	articleCollection, err := FindCollectionByID(userID)
	article, err := FindArticleByID(articleId)
	if err != nil {
		return err
	}
	article.StarNum++
	// 使用 Association 方法添加 Article 到 ArticleCollection
	if err := sqlDb.Model(&articleCollection).Association("ArticleCollection").Append(&article); err != nil {
		return err
	}
	err = UpdateArticle(article)
	if err != nil {
		return err
	}
	return nil
}

func FindCollectionByID(userId string) (model.FavoritesFolder, error) {
	var favoritesFolder model.FavoritesFolder
	if err := sqlDb.Where("user_id = ?", userId).First(&favoritesFolder).Error; err != nil {
		return model.FavoritesFolder{}, err //判断是不是空的
	}
	return favoritesFolder, nil
}

func AddCommentDB(userid string, newComment utils.AddCommentRequest) error {
	user, err := FindUserByUUID(userid)
	if err != nil {
		return err
	}
	comment := model.Comment{}
	now := time.Now()
	comment.CommentId = uuid.NewSHA1(uuid.NameSpaceDNS, []byte(now.String())).String()
	comment.SendUserId = userid
	comment.Content = newComment.Content
	comment.UserName = user.Name
	comment.UserAvatar = user.Avatar
	comment.PublishTime = time.Now()
	comment.LikedUsers = []model.User{}
	comment.RepliedComments = []model.Comment{}
	comment.ParentCommentId = nil
	comment.ArticleId = newComment.ArticleId
	if err := sqlDb.AutoMigrate(&comment); err != nil {
		return err
	}
	if err := sqlDb.Create(&comment).Error; err != nil {
		return err
	}
	if err := DeleteCommentCache(); err != nil {
		return err
	}
	article, err := FindArticleByID(newComment.ArticleId)
	if err != nil {
		return err
	}
	article.LikesNum++
	err = UpdateArticle(article)
	if err != nil {
		return err
	}
	return nil
}

// DeleteArticle sqlDb.Unscoped().Where("article_id = ?", id).Delete(&model.BloggerArticle{}).Error; err != nil Unscoped可以进行硬删除，彻底删除
// 简单一点 if err := sqlDb.Unscoped().Delete(&model.BloggerArticle{}, "article_id = ?", id).Error; err != nil {}这样也能删除
func DeleteArticle(id string) error {
	if err := sqlDb.Where("article_id = ?", id).Delete(&model.BloggerArticle{}).Error; err != nil {
		return err
	}
	return nil
}
func UpdateUser(user model.User) error {
	if err := sqlDb.Model(&model.User{}).Where("user_id = ?", user.UserId).Updates(user).Error; err != nil {
		return err
	}
	return nil
}

func RepliedCommentDb(request utils.RepliedCommentRequest) error {
	user, err := FindUserByUUID(request.SendUserId)
	if err != nil {
		return err
	}
	comment := model.Comment{}
	now := time.Now()
	comment.CommentId = uuid.NewSHA1(uuid.NameSpaceDNS, []byte(now.String())).String()
	comment.SendUserId = request.SendUserId
	comment.Content = request.Content
	comment.UserName = user.Name
	comment.UserAvatar = user.Avatar
	comment.PublishTime = time.Now()
	comment.LikedUsers = []model.User{}
	comment.RepliedComments = []model.Comment{}
	comment.ParentCommentId = &request.ParentID
	comment.ArticleId = request.ArticleId
	parentComment, err := FindCommentById(request.ParentID)
	if err != nil {
		return err
	}
	if err := sqlDb.Model(&parentComment).Association("RepliedComments").Append(&comment); err != nil {
		return err
	}
	return nil
}

func FindCommentById(commentId string) (model.Comment, error) {
	var comment model.Comment
	if err := sqlDb.Where("comment_id = ?", commentId).First(&comment).Error; err != nil {
		return comment, err
	}
	return comment, nil
}

func LikeCommentDB(commentId string, userId string) error {
	user, err := FindUserByUUID(userId)
	if err != nil {
		return err
	}
	comment, err := FindCommentById(commentId)
	if err != nil {
		return err
	}
	if err := sqlDb.Model(&comment).Association("LikedUsers").Append(&user); err != nil {
		return err
	}
	return nil
}

func GetCommentDB(articleId string, comments *[]model.Comment) error {
	if err := sqlDb.Where("article_id = ?", articleId).Find(&comments).Error; err != nil {
		return err
	}
	return nil
}

func SearchArticle(text string, page int, size int) (interface{}, interface{}, interface{}) {
	var articles []model.BloggerArticle
	var total int64
	total = int64(0)
	//计算符合条件的文章总数  %text% 	%表示任意字符序列  LIKE用于匹配符合条件的数据
	err := sqlDb.Model(&model.BloggerArticle{}).Where("title LIKE ?", "%"+text+"%").Count(&total).Error
	if err != nil {
		return articles, nil, err
	}
	//分页搜索，offset表示偏移，要第一页的内容就偏移0，第二页的内容就偏移1个size，以此类推，limit限制结果的大小
	err = sqlDb.Where("title LIKE ?", "%"+text+"%").Offset((page - 1) * size).Limit(size).Find(&articles).Error
	if err != nil {
		return articles, nil, err
	}
	return articles, total, nil
}

// GetCategory 数据库查询：
// 使用 Distinct() 方法确保查询结果是唯一的。
// 使用 Select("category") 指定只查询 category 字段。
// 使用 Scan(&results) 将查询结果映射到一个结构体切片中。 有点类似于.First，Scan输入，Value是获得值给变量
func GetCategory(results *[]string) error {
	if err := sqlDb.Model(&model.BloggerArticle{}).Distinct().Select("category").Scan(&results).Error; err != nil {
		return err
	}
	return nil
}

func GetArticlesByCategory(category string) ([]model.BloggerArticle, error) {
	var articles []model.BloggerArticle
	if err := sqlDb.Where("category = ?", category).Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

func CreateCustomizeFolderDB(name string, id string) (model.FavoritesFolder, error) {
	var folder model.FavoritesFolder
	folder.FolderName = name
	folder.UserId = id
	folder.ArticleCollection = []model.BloggerArticle{}
	now := time.Now()
	folder.FolderId = uuid.NewSHA1(uuid.NameSpaceDNS, []byte(now.String())).String()
	if err := sqlDb.Create(&folder).Error; err != nil {
		return model.FavoritesFolder{}, err
	}
	return folder, nil
}

func ModifyCustomizeFolder(folderId string, newName string) error {
	var folder model.FavoritesFolder
	if err := sqlDb.Where("folder_id = ?", folderId).First(&folder).Error; err != nil {
		return err
	}
	if folder.FolderName != newName {
		folder.FolderName = newName
		return errors.New("你使用了一样的名字")
	}
	if err := sqlDb.Model(&model.FavoritesFolder{}).Where("folder_id = ?", folderId).Updates(folder).Error; err != nil {
		return err
	}
	return nil
}

func GetAllFoldersDb(userId string) ([]model.FavoritesFolder, error) {
	var folders []model.FavoritesFolder
	if err := sqlDb.Where("user_id = ?", userId).Find(&folders).Error; err != nil {
		return []model.FavoritesFolder{}, err
	}
	return folders, nil
}
func GetFolderById(folderId string) (model.FavoritesFolder, error) {
	var folder model.FavoritesFolder
	if err := sqlDb.Where("folder_id = ?", folderId).First(&folder).Error; err != nil {
		return model.FavoritesFolder{}, err
	}
	return folder, nil
}

func GetArticleFromFolder(folderId string) ([]model.BloggerArticle, error) {
	var articles []model.BloggerArticle
	var folder model.FavoritesFolder
	if err := sqlDb.Preload("ArticleCollection").First(&folder, "folder_id = ?", folderId).Error; err != nil {
		return articles, err
	} //预加载folder相关联的ArticleCollection，把匹配的folderId文件夹赋值给folder，这样就能直接使用它的ArticleCollection
	articles = folder.ArticleCollection
	return articles, nil
}
