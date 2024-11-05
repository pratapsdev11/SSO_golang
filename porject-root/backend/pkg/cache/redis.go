package cache

import (
    "github.com/go-redis/redis/v8"
    "os"
)

func InitRedis() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     os.Getenv("REDIS_ADDR"),
        Password: os.Getenv("REDIS_PASSWORD"),
        DB:       0,
    })
}