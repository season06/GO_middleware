package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// 每10秒來自同一IP的請求數量不超過3次
var TIME_LIMITER = 10
var COUNT_LIMITER = 3

// redis: [key]:value -> [IP]:count
func RedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	pong, err := client.Ping().Result()
	log.Printf("Connect Redis Success !! ")
	log.Printf(pong, err)

	return client
}

func middleware(c *gin.Context) {
	client := RedisClient()

	IP := c.ClientIP()
	var ip_counter int
	var now_count int

	ip_counter, err := client.Get(IP).Int()
	if err == redis.Nil {
		log.Println("IP is not exist!!!!!!!!!")
		now_count = 1
		client.Set(IP, now_count, 10*time.Second)
	}

	ttl, _ := client.TTL(IP).Result()
	now_count = ip_counter + 1
	client.Set(IP, now_count, ttl)

	if ttl < time.Duration(0) {
		client.Del(IP)
	}

	log.Println("ip_counter: ", now_count)

	if now_count > COUNT_LIMITER {
		remain_time := strconv.Itoa(now_count - COUNT_LIMITER)
		c.Writer.Header().Set("X-RateLimit-Remaining", remain_time)
		c.Writer.Header().Set("X-RateLimit-Reset", ttl.String())

		c.String(http.StatusTooManyRequests, "http.StatusTooManyRequests")
		log.Panic("http.Status(429)")
		c.AbortWithStatus(429)
	}

	c.Next()
}

func output(c *gin.Context) {
	c.String(http.StatusOK, "hello, world")
}

func main() {
	router := gin.Default()
	router.GET("", middleware, output)
	router.Run(":8000")
}
