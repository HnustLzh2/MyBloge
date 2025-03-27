package db

import (
	"MyBloge/global"
	"github.com/go-redis/redis"
	"time"
)

func InitRedisValue() {
	redisDb = global.Redis
	articleCache = global.ArticleCacheKey
	commentCache = global.CommentCacheKey
}

var redisDb *redis.Client
var articleCache string
var commentCache string

func GetArticleFromRedis() (string, error) {
	cacheData, err := redisDb.Get(articleCache).Result()
	if err != nil {
		return "", err
	}
	return cacheData, nil
}
func SetArticleCache(articleByte []byte) error {
	if err := redisDb.Set(articleCache, articleByte, 20*time.Minute).Err(); err != nil {
		return err
	}
	return nil
}
func GetCommentCache() (string, error) {
	cacheData, err := redisDb.Get(commentCache).Result()
	DeleteCommentCache()
	if err != nil {
		return "", err
	}
	return cacheData, nil
}
func SetCommentCache(commentByte []byte) error {
	if err := redisDb.Set(commentCache, commentByte, 20*time.Minute).Err(); err != nil {
		return err
	}
	return nil
}

func DeleteArticleCache() error {
	if err := redisDb.Del(articleCache).Err(); err != nil {
		return err
	}
	return nil
}
func DeleteCommentCache() error {
	if err := redisDb.Del(commentCache).Err(); err != nil {
		return err
	}
	return nil
}
