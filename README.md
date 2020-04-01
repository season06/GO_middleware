# GO_middleware

## Installation & Run
```python
  # Download Gin Framework
  $ go get github.com/gin-gonic/gin
  
  # Download Redis Database
  $ go get github.com/go-redis/redis
```

## Require Parameters
  -> No more than 1,000 requests from the same IP per hour
```go
  var TIME_LIMITER = 3600
  var COUNT_LIMITER = 1000
```

## 1. Set Router
```go
  router := gin.Default()
  router.GET("", middleware, output)
  router.Run(":8000")
```

## 2. Set Redis Database Configuration
```go
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
```

## 3. Implement Limit Middleware
3-1.  Connet to the database.
```go
  client := RedisClient()
```
3-2.  Use function in gin framework "ClientIP()" to get the IP address.
```go
  IP := c.ClientIP()
```
3-3.  Redis has characteristics of 'key-value'.
   Let 'IP' as key, 'IP access times' as value.
   Use function in redis "Get" to check value of key.
   if 'error' presents the key does not exist, set the new key with expire time.
```go
var ip_counter int

ip_counter, err := client.Get(IP).Int()
if err == redis.Nil {
	log.Println("IP is not exist !!")
	counting = 1
	client.Set(IP, counting, 3600*time.Second)
}
```
4-4.  Use function in redis "TTL" to get the time in which the key will expire.
```go
ttl, _ := client.TTL(IP).Result()
counting = ip_counter + 1
client.Set(IP, counting, ttl)
```
Make sure if ttl == -1 or ttl == -2, delete this key.
```go
if ttl < time.Duration(0) {
	client.Del(IP)
}
```
4-5.  If the usage limit is exceeded, response http.statue(429) and show the information in the header.
```go
if counting > COUNT_LIMITER {
	remain_time := strconv.Itoa(counting - COUNT_LIMITER)
	c.Writer.Header().Set("X-RateLimit-Remaining", remain_time)
	c.Writer.Header().Set("X-RateLimit-Reset", ttl.String())

	c.String(http.StatusTooManyRequests, "Please don't come in too many times.\nThank you for your cooperation~")
	log.Panic("http.Status(429)")
	c.AbortWithStatus(429)
}
```
