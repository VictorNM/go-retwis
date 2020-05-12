package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type post struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

func (s *server) handlePostsCreate() echo.HandlerFunc {
	type response struct {
		ID int64 `json:"id"`
	}

	return func(c echo.Context) error {
		var req post
		if err := c.Bind(&req); err != nil {
			return rejectCode(c, http.StatusBadRequest)
		}

		curUserID, ok := getUserID(c)
		if !ok {
			return rejectCode(c, http.StatusNotFound)
		}

		p, err := createPost(s.client, curUserID, &post{Body: req.Body})
		if err != nil {
			return rejectCode(c, http.StatusInternalServerError)
		}

		return resolve(c, http.StatusCreated, response{ID: p.ID})
	}
}

func (s *server) handleGetUserPosts() echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := getUserID(c)
		if !ok {
			return rejectCode(c, http.StatusUnauthorized)
		}

		offset := int64(0)
		limit := int64(30)

		if c.QueryParam("offset") != "" {
			reqOffset, err := strconv.Atoi(c.QueryParam("offset"))
			if err != nil {
				return rejectCode(c, http.StatusBadRequest)
			}

			offset = int64(reqOffset)
		}
		if c.QueryParam("limit") != "" {
			reqLimit, err := strconv.Atoi(c.QueryParam("limit"))
			if err != nil {
				return rejectCode(c, http.StatusBadRequest)
			}

			limit = int64(reqLimit)
		}

		posts, err := getUserPosts(s.client, userID, offset, limit)
		if err != nil {
			return reject(c, http.StatusInternalServerError, err)
		}

		return resolve(c, http.StatusOK, posts)
	}
}