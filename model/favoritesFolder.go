package model

import (
	"gorm.io/gorm"
)

type FavoritesFolder struct {
	gorm.Model
	FolderId          string           `json:"id" gorm:"primaryKey;autoIncrement:false"`
	UserId            string           `json:"user_id" gorm:"foreignKey:user_id;references:UserId"`
	FolderName        string           `json:"folder_name"`
	ArticleCollection []BloggerArticle `json:"article_id_collection" gorm:"many2many:favorites_folder_article_id_collection;"`
	//不能使用切片类型，要么就直接使用切片对象，然后进行多对多处理，建立中间表
}
