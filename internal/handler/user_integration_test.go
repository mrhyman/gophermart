package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/mrhyman/gophermart/api"
	"github.com/mrhyman/gophermart/internal/handler"
	"github.com/mrhyman/gophermart/internal/repository"
	"github.com/mrhyman/gophermart/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*repository.Repository, func()) {

	if runtime.GOOS == "darwin" {
		t.Skip("Skipping integration test on macOS")
	}

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts("../../migrations/000001_user_table.up.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	repo, err := repository.NewRepository(connStr)
	require.NoError(t, err)

	cleanup := func() {
		repo.Close()
		pgContainer.Terminate(ctx)
	}

	return repo, cleanup
}

func TestUserHandler_Register_Integration(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	svc := service.New(repo)
	h := handler.New(*svc)

	t.Run("successful registration", func(t *testing.T) {
		reqBody := api.RegisterRequest{
			Login:    "newuser",
			Password: "password123",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		h.User.Register(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("duplicate registration returns conflict", func(t *testing.T) {
		login := "duplicateuser"
		reqBody := api.RegisterRequest{
			Login:    login,
			Password: "password123",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		// Первая регистрация
		req1 := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		rr1 := httptest.NewRecorder()
		h.User.Register(rr1, req1)
		assert.Equal(t, http.StatusOK, rr1.Code)

		// Повторная регистрация
		body2, _ := json.Marshal(reqBody)
		req2 := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body2))
		rr2 := httptest.NewRecorder()
		h.User.Register(rr2, req2)

		assert.Equal(t, http.StatusConflict, rr2.Code)
	})

	t.Run("empty login returns bad request", func(t *testing.T) {
		reqBody := api.RegisterRequest{
			Login:    "",
			Password: "password123",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.User.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("empty password returns bad request", func(t *testing.T) {
		reqBody := api.RegisterRequest{
			Login:    "userwithoupass",
			Password: "",
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.User.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid json returns bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte("invalid json")))
		rr := httptest.NewRecorder()

		h.User.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("wrong method returns method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/user/register", nil)
		rr := httptest.NewRecorder()

		h.User.Register(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

func TestUserHandler_Login_Integration(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	svc := service.New(repo)
	h := handler.New(*svc)

	// Подготовка: регистрируем пользователя
	login := "loginuser"
	password := "correctpassword"
	registerReq := api.RegisterRequest{
		Login:    login,
		Password: password,
	}
	body, _ := json.Marshal(registerReq)
	regReq := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
	regRR := httptest.NewRecorder()
	h.User.Register(regRR, regReq)
	require.Equal(t, http.StatusOK, regRR.Code)

	t.Run("successful login", func(t *testing.T) {
		loginReq := api.LoginRequest{
			Login:    login,
			Password: password,
		}
		body, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.User.Login(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("wrong password returns unauthorized", func(t *testing.T) {
		loginReq := api.LoginRequest{
			Login:    login,
			Password: "wrongpassword",
		}
		body, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.User.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("non-existent user returns unauthorized", func(t *testing.T) {
		loginReq := api.LoginRequest{
			Login:    "nonexistentuser",
			Password: "anypassword",
		}
		body, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.User.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("empty login returns bad request", func(t *testing.T) {
		loginReq := api.LoginRequest{
			Login:    "",
			Password: password,
		}
		body, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.User.Login(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("empty password returns bad request", func(t *testing.T) {
		loginReq := api.LoginRequest{
			Login:    login,
			Password: "",
		}
		body, err := json.Marshal(loginReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.User.Login(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid json returns bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader([]byte("invalid")))
		rr := httptest.NewRecorder()

		h.User.Login(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("wrong method returns method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/user/login", nil)
		rr := httptest.NewRecorder()

		h.User.Login(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}
