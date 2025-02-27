package model

import (
	"gorm.io/gorm"
)

type BloggerArticle struct {
	gorm.Model
	ArticleId   string `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Preview     string `json:"preview"`
	Category    string `json:"category"`
	StarNum     int    `json:"star_num" gorm:"default:0"`
	LikesNum    int    `json:"likes_num" gorm:"default:0"`
	CommentsNum int    `json:"comments_num" gorm:"default:0"`
	ViewNum     int    `json:"view_num" gorm:"default:0"`
	AuthorId    string `json:"author_id" gorm:"foreignKey:author_id;references:UserId"`
}
