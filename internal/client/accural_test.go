package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mrhyman/gophermart/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccrualClient_GetOrderAccrual_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orders/12345", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"order":"12345","status":"PROCESSED","accrual":500}`))
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	resp, err := client.GetOrderAccrual(context.Background(), "12345")

	require.NoError(t, err)
	assert.Equal(t, "12345", resp.Order)
	assert.Equal(t, model.AccrualStatusProcessed, resp.Status)
	assert.Equal(t, 500, *resp.Accrual)
}

func TestAccrualClient_GetOrderAccrual_OrderNotRegistered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	resp, err := client.GetOrderAccrual(context.Background(), "99999")

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, model.ErrOrderNotRegistered)
}

func TestAccrualClient_GetOrderAccrual_TooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Header().Set("Retry-After", "60")
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	resp, err := client.GetOrderAccrual(context.Background(), "12345")

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, model.ErrAccrualTooManyRequests)
}

func TestAccrualClient_GetOrderAccrual_InternalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	resp, err := client.GetOrderAccrual(context.Background(), "12345")

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, model.ErrAccrualInternalError)
}

func TestAccrualClient_GetOrderAccrual_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := NewAccrualClient(server.URL)

	resp, err := client.GetOrderAccrual(context.Background(), "12345")

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with http", "http://localhost:8080", "http://localhost:8080"},
		{"with https", "https://example.com", "https://example.com"},
		{"without scheme", "localhost:8080", "http://localhost:8080"},
		{"domain only", "example.com", "http://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeBaseURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
