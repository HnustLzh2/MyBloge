package global

import (
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var (
	SqlDb *gorm.DB
	Redis *redis.Client
)

var ArticleCacheKey = "Articles"
var CommentCacheKey = "Comments"
