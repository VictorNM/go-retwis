package main

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"net/http"
	"reflect"
	"strconv"
)

func (s *server) handleFollow() echo.HandlerFunc {
	type response struct {
		Following bool `json:"following"`
	}

	return func(c echo.Context) error {
		token, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return reject(c, http.StatusInternalServerError, errors.New(reflect.TypeOf(c.Get("user")).Elem().Name()))
		}

		claims, ok := token.Claims.(*jwtClaims)
		if !ok {
			return reject(c, http.StatusInternalServerError, errors.New(reflect.TypeOf(token.Claims).Elem().Name()))
		}

		followingID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return rejectCode(c, http.StatusBadRequest)
		}

		if isFollowing(s.client, int64(followingID), claims.UserID) {
			err = deleteFollow(s.client, int64(followingID), claims.UserID)
			if err != nil {
				return reject(c, http.StatusInternalServerError, err)
			}

			return resolve(c, http.StatusOK, response{Following: false})
		}

		err = createFollow(s.client, int64(followingID), claims.UserID)
		if err != nil {
			return reject(c, http.StatusInternalServerError, err)
		}

		return resolve(c, http.StatusOK, response{Following: true})
	}
}
