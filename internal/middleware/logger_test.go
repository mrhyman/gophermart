package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mrhyman/gophermart/internal/logger"
	"github.com/mrhyman/gophermart/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestWithLogging(t *testing.T) {
	t.Run("log successful request", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		log := logger.NewWithCore(core)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("response body"))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := logger.WithinContext(req.Context(), log)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		loggedHandler := WithLogging(handler)
		loggedHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, 1, recorded.Len())

		logEntry := recorded.All()[0]
		assert.Equal(t, zapcore.InfoLevel, logEntry.Level)

		contextMap := logEntry.ContextMap()
		assert.Equal(t, "/test", contextMap["uri"])
		assert.Equal(t, http.MethodGet, contextMap["method"])
		assert.Equal(t, int64(http.StatusOK), contextMap["status"])
		assert.Equal(t, int64(13), contextMap["size"]) // len("response body")
		assert.Contains(t, contextMap, "duration")
	})

	t.Run("log request with chi route pattern", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		log := logger.NewWithCore(core)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		r := chi.NewRouter()
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := logger.WithinContext(r.Context(), log)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
		r.Get("/users/{id}", WithLogging(handler))

		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, 1, recorded.Len())
		contextMap := recorded.All()[0].ContextMap()
		assert.Equal(t, "/users/{id}", contextMap["uri"])
	})

	t.Run("log error status", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		log := logger.NewWithCore(core)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("error"))
		})

		req := httptest.NewRequest(http.MethodPost, "/error", nil)
		ctx := logger.WithinContext(req.Context(), log)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		loggedHandler := WithLogging(handler)
		loggedHandler.ServeHTTP(rr, req)

		assert.Equal(t, 1, recorded.Len())
		contextMap := recorded.All()[0].ContextMap()
		assert.Equal(t, int64(http.StatusBadRequest), contextMap["status"])
		assert.Equal(t, int64(5), contextMap["size"])
	})

	t.Run("log request with userID in context", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		log := logger.NewWithCore(core)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		ctx := logger.WithinContext(req.Context(), log)
		ctx = context.WithValue(ctx, model.UserIDKey, "user-123")
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		loggedHandler := WithLogging(handler)
		loggedHandler.ServeHTTP(rr, req)

		assert.Equal(t, 1, recorded.Len())
		contextMap := recorded.All()[0].ContextMap()
		assert.Equal(t, "user-123", contextMap["userID"])
	})

	t.Run("log request without userID", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		log := logger.NewWithCore(core)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/public", nil)
		ctx := logger.WithinContext(req.Context(), log)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		loggedHandler := WithLogging(handler)
		loggedHandler.ServeHTTP(rr, req)

		assert.Equal(t, 1, recorded.Len())
		contextMap := recorded.All()[0].ContextMap()
		assert.Nil(t, contextMap["userID"])
	})

	t.Run("log empty response", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		log := logger.NewWithCore(core)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodDelete, "/resource", nil)
		ctx := logger.WithinContext(req.Context(), log)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		loggedHandler := WithLogging(handler)
		loggedHandler.ServeHTTP(rr, req)

		assert.Equal(t, 1, recorded.Len())
		contextMap := recorded.All()[0].ContextMap()
		assert.Equal(t, int64(http.StatusNoContent), contextMap["status"])
		assert.Equal(t, int64(0), contextMap["size"])
	})

	t.Run("measure duration", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		log := logger.NewWithCore(core)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/slow", nil)
		ctx := logger.WithinContext(req.Context(), log)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		loggedHandler := WithLogging(handler)
		loggedHandler.ServeHTTP(rr, req)

		assert.Equal(t, 1, recorded.Len())
		contextMap := recorded.All()[0].ContextMap()
		duration := contextMap["duration"].(string)
		assert.NotEmpty(t, duration)
		assert.Contains(t, duration, "ms")
	})

	t.Run("log multiple writes", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		log := logger.NewWithCore(core)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("part1"))
			w.Write([]byte("part2"))
			w.Write([]byte("part3"))
		})

		req := httptest.NewRequest(http.MethodGet, "/multi", nil)
		ctx := logger.WithinContext(req.Context(), log)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		loggedHandler := WithLogging(handler)
		loggedHandler.ServeHTTP(rr, req)

		assert.Equal(t, 1, recorded.Len())
		contextMap := recorded.All()[0].ContextMap()
		assert.Equal(t, int64(15), contextMap["size"]) // 5+5+5
	})
}

func TestLogWriter(t *testing.T) {
	t.Run("write sets status to OK if not set", func(t *testing.T) {
		rr := httptest.NewRecorder()
		lw := &logWriter{ResponseWriter: rr}

		n, err := lw.Write([]byte("test"))

		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, http.StatusOK, lw.status)
		assert.Equal(t, 4, lw.size)
	})

	t.Run("writeHeader sets status once", func(t *testing.T) {
		rr := httptest.NewRecorder()
		lw := &logWriter{ResponseWriter: rr}

		lw.WriteHeader(http.StatusCreated)
		lw.WriteHeader(http.StatusBadRequest) // не должно измениться

		assert.Equal(t, http.StatusCreated, lw.status)
	})

	t.Run("multiple writes accumulate size", func(t *testing.T) {
		rr := httptest.NewRecorder()
		lw := &logWriter{ResponseWriter: rr}

		lw.Write([]byte("hello"))
		lw.Write([]byte(" "))
		lw.Write([]byte("world"))

		assert.Equal(t, 11, lw.size)
	})

	t.Run("writeHeader before write", func(t *testing.T) {
		rr := httptest.NewRecorder()
		lw := &logWriter{ResponseWriter: rr}

		lw.WriteHeader(http.StatusAccepted)
		lw.Write([]byte("data"))

		assert.Equal(t, http.StatusAccepted, lw.status)
		assert.Equal(t, 4, lw.size)
	})
}
