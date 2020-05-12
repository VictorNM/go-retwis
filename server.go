package main

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"net/http"
)

type server struct {
	e *echo.Echo

	client *redis.Client

	jwtSecret string
}

func newServer(e *echo.Echo, client *redis.Client) *server {
	e.Validator = &myValidator{v: validator.New()}
	s := &server{
		e:      e,
		client: client,
	}

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.e.ServeHTTP(w, r)
}

func (s *server) routes() {
	api := s.e.Group("/api")
	api.Use(s.jwtMiddleware("/api"))

	api.POST("/auth/register", s.authRegister())
	api.POST("/auth/login", s.authLogin())

	// authenticated handler
	api.POST("/follows/:id", s.handleFollow())
	api.GET("/users/me", s.handleUsersMe())
	api.GET("/users/:id", s.handleUsersProfile())
	api.POST("/posts", s.handlePostsCreate())
	api.GET("/posts", s.handleGetUserPosts())
}

func getUserID(c echo.Context) (int64, bool) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return 0, false
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return 0, false
	}

	return claims.UserID, true
}

func resolve(c echo.Context, code int, data interface{}) error {
	return c.JSON(code, BaseResponse{Data: data})
}

func reject(c echo.Context, code int, err error) error {
	return c.JSON(code, BaseResponse{
		Error: BaseError{
			Message: err.Error(),
		},
	})
}

func rejectCode(c echo.Context, code int) error {
	return reject(c, code, errors.New(http.StatusText(code)))
}

type BaseError struct {
	Message string `json:"message,omitempty"`
}

type BaseResponse struct {
	Error BaseError   `json:"error"`
	Data  interface{} `json:"data,omitempty"`
}
