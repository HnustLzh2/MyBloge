package model

import (
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	gorm.Model
	CommentId       string    `json:"id" gorm:"primaryKey;autoIncrement:false"`
	SendUserId      string    `json:"send_userid" gorm:"foreignKey:send_userid;reference:UserId"` //发送者iD
	UserAvatar      string    `json:"user_avatar"`
	UserName        string    `json:"user_name"`
	PublishTime     time.Time `json:"publish_time"`
	Content         string    `gorm:"type:text" json:"content"`
	LikeCount       int       `json:"like_count"`
	RepliedComments []Comment `json:"replied_comments_id" gorm:"foreignKey:ParentCommentId;references:CommentId"` //回复的评论id号
	ParentCommentId string    `json:"parent_comment_id" gorm:"foreignKey:Id;references:CommentId"`
	LikeUserId      []string  `json:"like_user_id"` //点赞的人群的Id集合
}
