package main

import (
	"flag"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"os"
	"strconv"
)

type myValidator struct {
	v *validator.Validate
}

func (v *myValidator) Validate(i interface{}) error {
	return v.v.Struct(i)
}

func main() {
	var (
		httpAddr  = strEnv("HTTP_ADDR", ":8080")
		redisAddr = strEnv("REDIS_ADDR", ":6379")
		redisPass = strEnv("REDIS_PASS", "")
		redisDB   = intEnv("REDIS_DB", 0)
		jwtSecret = strEnv("JWT_SECRET", "")
	)

	httpAddr 	= *flag.String("http-addr", httpAddr, "address for serving HTTP")
	redisAddr 	= *flag.String("redis-addr", redisAddr, "address for redis")
	redisPass	= *flag.String("redis-pass", "", "password for redis")
	redisDB		= *flag.Int("redis-db", 0, "redis database index")
	jwtSecret 	= *flag.String("jwt-secret", jwtSecret, "secret key for jwt")

	flag.Parse()

	e := echo.New()
	redisOpts := &redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       redisDB,
	}

	client := redis.NewClient(redisOpts)

	s := newServer(e, client)
	s.jwtSecret = jwtSecret
	s.routes()

	e.Logger.Fatal(s.e.Start(httpAddr))
}

func intEnv(name string, value int) int {
	if v, e := strconv.Atoi(os.Getenv(name)); e == nil {
		return v
	}

	return value
}

func strEnv(name string, value string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}

	return value
}
