package global

import (
	"MyBloge/websockets"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var (
	SqlDb           *gorm.DB
	Redis           *redis.Client
	GlobalPool      *websockets.Pool
	ArticleCacheKey string
	CommentCacheKey string
)

func init() {
	ArticleCacheKey = "articleCache"
	CommentCacheKey = "commentCache"
}
