package main

import (
	"encoding/json"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victornm/jat"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var client *redis.Client

func testServer() *server {
	e := echo.New()
	e.Use(middleware.Logger())
	s := newServer(e, client)

	return s
}

func TestRegister(t *testing.T) {
	defer resetDB()

	s := testServer()
	s.routes()

	w := performRegister(s, "victor", "1234abcd")

	assert.Equal(t, 201, w.Code)

	data := getDataAsMap(t, w)
	assert.NotEmpty(t, data["token"])
}

func TestLogin(t *testing.T) {
	defer resetDB()

	s := testServer()
	s.routes()

	performRegister(s, "victor", "1234abcd")
	w := performLogin(s, "victor", "1234abcd")

	assert.Equal(t, 200, w.Code)

	data := getDataAsMap(t, w)
	assert.NotEmpty(t, data["token"])
}

func TestFollow(t *testing.T) {
	defer resetDB()

	s := testServer()
	s.routes()

	performRegister(s, "victor", "1234abcd")
	w := performRegister(s, "victor_follower", "1234abcd")

	token := getDataAsMap(t, w)["token"].(string)

	victor, err := getUserByUsername(client, "victor")
	require.NoError(t, err)

	w = performFollow(s, token, victor.id)
	follower, err := getUserByUsername(client, "victor_follower")
	require.NoError(t, err)

	assert.Equal(t, 200, w.Code)
	assert.True(t, isFollowing(client, victor.id, follower.id))

	// test unfollow
	w = performFollow(s, token, victor.id)
	assert.Equal(t, 200, w.Code)
	assert.False(t, isFollowing(client, victor.id, follower.id))
}

func TestGetMe(t *testing.T) {
	defer resetDB()

	s := testServer()
	s.routes()

	w := performRegister(s, "victor", "1234abcd")

	token := getDataAsMap(t, w)["token"].(string)

	w = performGetMe(s, token)

	assert.Equal(t, 200, w.Code)
	data := getDataAsMap(t, w)
	assert.Equal(t, "victor", data["username"])
}

func TestGetProfile(t *testing.T) {
	defer resetDB()

	s := testServer()
	s.routes()

	performRegister(s, "victor", "1234abcd")
	w := performRegister(s, "bruce", "1234abcd")

	token := getDataAsMap(t, w)["token"].(string)

	victor, err := getUserByUsername(client, "victor")
	require.NoError(t, err)
	w = performGetProfile(s, token, victor.id)

	assert.Equal(t, 200, w.Code)
	data := getDataAsMap(t, w)
	assert.Equal(t, "victor", data["username"])
}

func performRegister(handler http.Handler, username, password string) *httptest.ResponseRecorder {
	r := jat.WrapPOST("/api/auth/register", map[string]string{
		"username": username,
		"password": password,
	}).
		SetHeader(echo.HeaderContentType, echo.MIMEApplicationJSON).
		Unwrap()
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)
	return w
}

func performLogin(handler http.Handler, username, password string) *httptest.ResponseRecorder {
	r := jat.WrapPOST("/api/auth/login", nil).
		SetBasicAuth(username, password).
		SetHeader(echo.HeaderContentType, echo.MIMEApplicationJSON).
		Unwrap()
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	return w
}

func performFollow(handler http.Handler, token string, id int64) *httptest.ResponseRecorder {
	r := jat.WrapPOST("/api/follows/:id", nil).
		SetBearerAuth(token).
		SetParam("id", id).
		SetHeader(echo.HeaderContentType, echo.MIMEApplicationJSON).
		Unwrap()
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	return w
}

func performGetMe(handler http.Handler, token string) *httptest.ResponseRecorder {
	r := jat.WrapGET("/api/users/me").
		SetBearerAuth(token).
		SetHeader(echo.HeaderContentType, echo.MIMEApplicationJSON).
		Unwrap()
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	return w
}

func performGetProfile(handler http.Handler, token string, id int64) *httptest.ResponseRecorder {
	r := jat.WrapGET("/api/users/:id").
		SetParam("id", id).
		SetBearerAuth(token).
		SetHeader(echo.HeaderContentType, echo.MIMEApplicationJSON).
		Unwrap()
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	return w
}

func resetDB() {
	client.FlushDB()
}

func getData(t *testing.T, w *httptest.ResponseRecorder) interface{} {
	t.Helper()

	var res BaseResponse

	err := json.Unmarshal([]byte(w.Body.String()), &res)
	if err != nil {
		return nil
	}

	return res.Data
}

func getDataAsMap(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()

	data, ok := getData(t, w).(map[string]interface{})
	if !ok {
		log.Panicf("response data is not a %T", map[string]interface{}{})
	}

	return data
}

func TestMain(m *testing.M) {
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	defer func() {
		client.FlushDB()
		_ = client.Close()
	}()

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("ping redis client failed: %v", err)
	}

	os.Exit(m.Run())
}
