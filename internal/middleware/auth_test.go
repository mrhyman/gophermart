package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrhyman/gophermart/internal/auth"
	"github.com/mrhyman/gophermart/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithAuth(t *testing.T) {
	secret := "test-secret"

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(model.UserIDKey)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userID.(string)))
	})

	t.Run("success with valid cookie", func(t *testing.T) {
		ce, err := auth.NewCookieEncoder(secret)
		require.NoError(t, err)

		userID := "test-user-123"
		encodedUserID, err := ce.EncodeUserID(userID)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "X-Auth-Token",
			Value: encodedUserID,
		})

		rr := httptest.NewRecorder()

		handler := WithAuth(secret)(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, userID, rr.Body.String())
	})

	t.Run("fail when cookie is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler := WithAuth(secret)(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), model.ErrWentWrong.Error())
	})

	t.Run("fail with invalid cookie value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "X-Auth-Token",
			Value: "invalid-cookie-value",
		})

		rr := httptest.NewRecorder()

		handler := WithAuth(secret)(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), model.ErrUnknownUser.Error())
	})

	t.Run("fail with wrong secret", func(t *testing.T) {
		ce, err := auth.NewCookieEncoder("different-secret")
		require.NoError(t, err)

		userID := "test-user"
		encodedUserID, err := ce.EncodeUserID(userID)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "X-Auth-Token",
			Value: encodedUserID,
		})

		rr := httptest.NewRecorder()

		handler := WithAuth(secret)(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), model.ErrUnknownUser.Error())
	})

	t.Run("context contains userID after middleware", func(t *testing.T) {
		ce, err := auth.NewCookieEncoder(secret)
		require.NoError(t, err)

		expectedUserID := "context-test-user"
		encodedUserID, err := ce.EncodeUserID(expectedUserID)
		require.NoError(t, err)

		var capturedUserID string
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(model.UserIDKey).(string)
			require.True(t, ok, "userID should be string")
			capturedUserID = userID
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "X-Auth-Token",
			Value: encodedUserID,
		})

		rr := httptest.NewRecorder()

		handler := WithAuth(secret)(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, expectedUserID, capturedUserID)
	})
}
