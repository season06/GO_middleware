package main

import (
	"log"
	"net/http"
	"reflect"
	"strconv"

	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

// 每5秒來自同一IP的請求數量不超過3次
var time_limiter = 5
var count_limiter = 3

// redis: [key]:value -> [IP]:count
// var client redis.Conn

func getVisitor(client redis.Conn, IP string) int {
	exist, err := redis.Bool(client.Do("EXISTS", IP))
	if err != nil {
		log.Panic("client EXISTS error: ", err)
	}
	log.Println("exist ", exist)
	log.Println("IP ", IP)

	if !exist {
		log.Println("IP is not exist!!!!!!!!!")
		_, err = client.Do("SET", IP, 1, "EX", "3600") // after 3600s key will expire
		if err != nil {
			log.Panic("redis SET failed:", err)
		}
		return 1
	}
	var val, _ = client.Do("GET", IP)
	log.Println("val_type: ", reflect.TypeOf(val))
	log.Println("val: ", val)
	// log.Println("val: ", strconv.Atoi(val.(string)))
	var new_val = val. + 1
	client.Do("GETSET", IP, new_val)

	log.Println(new_val)
	return int(new_val)
}

func middleware(c *gin.Context) {
	client, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Panic("Connect to redis error", err)
	}
	log.Println("Connect Redis Success !!!!!!")
	// defer client.Close()

	IP := c.ClientIP()
	var request_count = getVisitor(client, IP)
	log.Println("count : ", request_count)

	if request_count > count_limiter {
		var remain_time = strconv.Itoa(request_count - count_limiter)
		// reply, err := client.Do("TLL", IP)
		// if err != nil {
		// 	log.Panic("get IP 'TLL' error: ", err)
		// } else {
		// 	log.Println(reply)
		// 	// var reset_time = strconv.Itoa(reply)
		// }
		c.Writer.Header().Set("X-RateLimit-Remaining", remain_time)
		c.Writer.Header().Set("X-RateLimit-Reset", remain_time)
		// log.Println("X-RateLimit-Remaining: ", c.Request.Header["X-RateLimit-Remaining"])
		// log.Println("X-RateLimit-Reset: ", c.Request.Header["X-RateLimit-Reset"])

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
