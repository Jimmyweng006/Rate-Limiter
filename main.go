package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

func newRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return client
}

func RateLimiter(redisClient *redis.Client, requestLimit int, requestResetTime int) gin.HandlerFunc {
	return func(c *gin.Context) {
		remainRequest, err := redisClient.Get(c.ClientIP()).Result()

		if err != nil {
			err := redisClient.Set(c.ClientIP(), requestLimit-1, time.Duration(requestResetTime)*time.Second).Err()
			if err != nil {
				panic(err)
			}
		} else {
			redisClient.Decr(c.ClientIP())
			remainRequest, err = redisClient.Get(c.ClientIP()).Result()
			expire, _ := redisClient.TTL(c.ClientIP()).Result()

			if remainRequestInt, _ := strconv.Atoi(remainRequest); remainRequestInt < 0 {
				c.Abort()

				c.JSON(http.StatusTooManyRequests, gin.H{
					"request reset time": time.Unix(int64(expire.Seconds())+time.Now().Unix(), 0),
				})
			}
		}

		c.Next()
	}
}

func main() {
	r := gin.Default()
	redisClient := newRedisClient()
	requestLimit := 3
	requestResetTime := 60
	r.Use(RateLimiter(redisClient, requestLimit, requestResetTime))

	r.GET("/", func(c *gin.Context) {
		remainRequest, _ := redisClient.Get(c.ClientIP()).Result()
		expire, _ := redisClient.TTL(c.ClientIP()).Result()

		c.Header("X-RateLimit-Remaining", remainRequest)
		c.Header("X-RateLimit-Reset", time.Unix(int64(expire.Seconds())+time.Now().Unix(), 0).Format("2006-01-02 15:04:05"))
		c.JSON(http.StatusOK, gin.H{
			"hello":                 "word",
			"X-RateLimit-Remaining": remainRequest,
		})
	})

	r.Run()
}
