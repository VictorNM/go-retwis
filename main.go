package main

import (
	"flag"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
)

type myValidator struct {
	v *validator.Validate
}

func (v *myValidator) Validate(i interface{}) error {
	return v.v.Struct(i)
}

func main() {
	var (
		httpAddr string

		redisAddr string
		redisPass string
		redisDB   int

		jwtSecret string
	)

	flag.StringVar(&httpAddr, "http-addr", ":8080", "address for serving HTTP")
	flag.StringVar(&redisAddr, "redis-addr", ":6379", "address for redis")
	flag.StringVar(&redisPass, "redis-pass", "", "password for redis")
	flag.IntVar(&redisDB, "redis-db", 0, "redis database index")
	flag.StringVar(&jwtSecret, "jwt-secret", "", "secret key for jwt")

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
