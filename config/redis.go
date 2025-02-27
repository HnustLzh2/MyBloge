package config

import (
	"MyBloge/global"
	"github.com/go-redis/redis"
	"log"
)

func InitRedis() {
	RedisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := RedisClient.Ping().Result()
	if err != nil {
		log.Fatalf("redis connect err: %v", err)
	}
	global.Redis = RedisClient
}
