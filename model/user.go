package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserId            string `json:"id" gorm:"primaryKey;autoIncrement:false"` //向外暴露的标识符
	Name              string `json:"name"`
	Authorization     string `json:"authorization"`
	Password          string `json:"password"`
	Email             string `json:"email"`
	Avatar            string `json:"avatar"`                     //头像
	FollowNum         int    `json:"followNum" gorm:"default:0"` //关注数
	FansNum           int    `json:"fansNum" gorm:"default:0"`   //粉丝数
	FavoritesFolderId string `json:"favorites_folder_id" gorm:"foreignKey:favorites_folder_id;references:FolderId"`
	Token             string `json:"tokens"`
	RefreshToken      string `json:"refresh_token"`
}
