package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// 每1hr來自同一IP的請求數量不超過1000次
var TIME_LIMITER = 3600
var COUNT_LIMITER = 1000
var counting int

// redis: [key]:value -> [IP]:count
func RedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	pong, err := client.Ping().Result()
	log.Printf("Connect Redis Success !!")
	log.Printf(pong, err)

	return client
}

func middleware(c *gin.Context) {
	client := RedisClient()

	IP := c.ClientIP()
	var ip_counter int
	// var counting int

	ip_counter, err := client.Get(IP).Int()
	if err == redis.Nil {
		log.Println("IP is not exist !!")
		counting = 1
		client.Set(IP, counting, 3600*time.Second)
	}

	ttl, _ := client.TTL(IP).Result()
	counting = ip_counter + 1
	client.Set(IP, counting, ttl)

	if ttl < time.Duration(0) {
		client.Del(IP)
	}

	log.Println("ip_counter: ", counting)

	if counting > COUNT_LIMITER {
		remain_time := strconv.Itoa(counting - COUNT_LIMITER)
		c.Writer.Header().Set("X-RateLimit-Remaining", remain_time)
		c.Writer.Header().Set("X-RateLimit-Reset", ttl.String())

		c.String(http.StatusTooManyRequests, "Please don't come in too many times.\nThank you for your cooperation~")
		log.Panic("http.Status(429)")
		c.AbortWithStatus(429)
	}

	c.Next()
}

func output(c *gin.Context) {
	c.String(http.StatusOK, fmt.Sprintf("Thank you for your click~\nThis is the times you clicked: %d", counting))
}

func main() {
	router := gin.Default()
	router.GET("", middleware, output)
	router.Run(":8000")
}
