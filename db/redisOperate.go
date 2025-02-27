package db

import (
	"MyBloge/global"
	"time"
)

var redisDb = global.Redis
var articleCache = global.ArticleCacheKey
var commentCache = global.CommentCacheKey

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
