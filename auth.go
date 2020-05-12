package main

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"path"
	"strconv"
	"time"
)

type user struct {
	id       int64
	username string
	password string
}

func (s *server) authRegister() echo.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type response struct {
		Token string `json:"token"`
	}

	return func(c echo.Context) error {
		var req request
		if err := c.Bind(&req); err != nil {
			return reject(c, http.StatusBadRequest, err)
		}

		if err := c.Validate(req); err != nil {
			return reject(c, http.StatusBadRequest, err)
		}

		_, err := getUserByUsername(s.client, req.Username)
		if err == nil {
			return reject(c, http.StatusBadRequest, errors.New("username existed"))
		}

		user := &user{username: req.Username, password: req.Password}
		user, err = createUser(s.client, user)
		if err != nil {
			return rejectCode(c, http.StatusInternalServerError)
		}

		token, err := s.genToken(user.id)
		if err != nil {
			return reject(c, http.StatusInternalServerError, err)
		}

		return resolve(c, http.StatusCreated, response{Token: token})
	}
}

func (s *server) authLogin() echo.HandlerFunc {
	type response struct {
		Token string `json:"token"`
	}

	return func(c echo.Context) error {
		username, password, ok := c.Request().BasicAuth()
		if !ok {
			return rejectCode(c, http.StatusUnauthorized)
		}

		user, err := getUserByUsername(s.client, username)
		if err != nil {
			return reject(c, http.StatusUnauthorized, err)
		}

		if user.password != password {
			return reject(c, http.StatusUnauthorized, err)
		}

		token, err := s.genToken(user.id)
		if err != nil {
			return reject(c, http.StatusInternalServerError, err)
		}

		return resolve(c, http.StatusOK, response{Token: token})
	}
}

type jwtClaims struct {
	jwt.StandardClaims
	UserID int64 `json:"user_id"`
}

func (s *server) genToken(userID int64) (string, error) {
	expiredDuration := time.Hour * 60
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwtClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiredDuration).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "auth",
		},
		UserID: userID,
	})

	log.Println(userID)

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("generate JWT token failed: %v", err)
	}

	return tokenString, nil
}

func (s *server) jwtMiddleware(basePath string) echo.MiddlewareFunc {
	return middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:    []byte(s.jwtSecret),
		SigningMethod: jwt.SigningMethodHS256.Name,
		ContextKey:    "user",
		Claims:        &jwtClaims{},
		TokenLookup:   "header:Authorization",
		AuthScheme:    "Bearer",
		Skipper: func(c echo.Context) bool {
			if c.Path() == path.Join(basePath, "auth", "register") || c.Path() == path.Join(basePath, "auth", "login") {
				return true
			}

			return false
		},
	})
}


func (s *server) handleUsersMe() echo.HandlerFunc {
	type response struct {
		ID             int64  `json:"id"`
		Username       string `json:"username"`
		FollowersCount int    `json:"followers_count"`
		FollowingCount int    `json:"following_count"`
	}

	return func(c echo.Context) error {
		userID, ok := getUserID(c)
		if !ok {
			return rejectCode(c, http.StatusUnauthorized)
		}

		user, err := getUserByID(s.client, userID)
		if err != nil {
			return rejectCode(c, http.StatusUnauthorized)
		}

		res := response{
			ID:             user.id,
			Username:       user.username,
			FollowersCount: countFollowers(s.client, user.id),
			FollowingCount: countFollowing(s.client, user.id),
		}

		return resolve(c, http.StatusOK, res)
	}
}

func (s *server) handleUsersProfile() echo.HandlerFunc {
	type response struct {
		ID          int64  `json:"id"`
		Username    string `json:"username"`
		IsFollowing bool   `json:"is_following"`
	}

	return func(c echo.Context) error {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return rejectCode(c, http.StatusBadRequest)
		}

		user, err := getUserByID(s.client, int64(userID))
		if err != nil {
			return rejectCode(c, http.StatusNotFound)
		}

		curUserID, ok := getUserID(c)
		if !ok {
			return rejectCode(c, http.StatusUnauthorized)
		}

		res := response{
			ID:          user.id,
			Username:    user.username,
			IsFollowing: isFollowing(s.client, int64(userID), curUserID),
		}

		return resolve(c, http.StatusOK, res)
	}
}