package model

import "time"

// Comment 1对多关系不需要建立中间表，但是需要指定外键***
type Comment struct {
	CommentId       string    `json:"id" gorm:"primaryKey;autoIncrement:false"`
	SendUserId      string    `json:"send_userid" gorm:"foreignKey:SendUserId;references:UserId"`
	UserAvatar      string    `json:"user_avatar"`
	UserName        string    `json:"user_name"`
	PublishTime     time.Time `json:"publish_time"`
	Content         string    `gorm:"type:text" json:"content"`
	LikeCount       int       `json:"like_count"`
	ArticleId       string    `json:"article_id" gorm:"foreignKey:ArticleId;references:ArticleId"`
	ParentCommentId *string   `json:"parent_comment_id" gorm:"index"`
	RepliedComments []Comment `json:"replied_comments" gorm:"foreignKey:ParentCommentId;references:CommentId"`
	LikedUsers      []User    `gorm:"many2many:comment_likes;" json:"liked_users"`
}
