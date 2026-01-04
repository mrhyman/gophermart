package auth

import (
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCookieEncoder(t *testing.T) {
	t.Run("success with valid secret", func(t *testing.T) {
		encoder, err := NewCookieEncoder("my-secret")

		require.NoError(t, err)
		assert.NotNil(t, encoder)
		assert.NotNil(t, encoder.crypto)
	})

	t.Run("success with empty secret", func(t *testing.T) {
		encoder, err := NewCookieEncoder("")

		require.NoError(t, err)
		assert.NotNil(t, encoder)
	})
}

func TestCookieEncoder_EncodeUserID(t *testing.T) {
	encoder, err := NewCookieEncoder("test-secret")
	require.NoError(t, err)

	t.Run("encode valid UUID", func(t *testing.T) {
		userID := uuid.New().String()

		encoded, err := encoder.EncodeUserID(userID)

		require.NoError(t, err)
		assert.NotEmpty(t, encoded)
		assert.NotEqual(t, userID, encoded)
		// Проверяем, что это валидный hex
		_, hexErr := hex.DecodeString(encoded)
		assert.NoError(t, hexErr)
	})

	t.Run("encode simple string", func(t *testing.T) {
		userID := "user123"

		encoded, err := encoder.EncodeUserID(userID)

		require.NoError(t, err)
		assert.NotEmpty(t, encoded)
	})

	t.Run("encode empty string", func(t *testing.T) {
		encoded, err := encoder.EncodeUserID("")

		require.NoError(t, err)
		assert.NotEmpty(t, encoded) // nonce всегда есть
	})

	t.Run("different encodings for same userID", func(t *testing.T) {
		userID := "same-user-id"

		encoded1, err1 := encoder.EncodeUserID(userID)
		encoded2, err2 := encoder.EncodeUserID(userID)

		require.NoError(t, err1)
		require.NoError(t, err2)
		// Из-за разных nonce результаты должны отличаться
		assert.NotEqual(t, encoded1, encoded2)
	})
}

func TestCookieEncoder_DecodeUserID(t *testing.T) {
	encoder, err := NewCookieEncoder("test-secret")
	require.NoError(t, err)

	t.Run("decode valid cookie", func(t *testing.T) {
		userID := uuid.New().String()
		encoded, err := encoder.EncodeUserID(userID)
		require.NoError(t, err)

		cookie := &http.Cookie{Value: encoded}

		decoded, err := encoder.DecodeUserID(cookie)

		require.NoError(t, err)
		assert.Equal(t, userID, decoded)
	})

	t.Run("decode empty userID", func(t *testing.T) {
		encoded, err := encoder.EncodeUserID("")
		require.NoError(t, err)

		cookie := &http.Cookie{Value: encoded}

		decoded, err := encoder.DecodeUserID(cookie)

		require.NoError(t, err)
		assert.Empty(t, decoded)
	})

	t.Run("fail on invalid hex", func(t *testing.T) {
		cookie := &http.Cookie{Value: "not-a-hex-string!!!"}

		_, err := encoder.DecodeUserID(cookie)

		assert.Error(t, err)
	})

	t.Run("fail on corrupted data", func(t *testing.T) {
		userID := "test-user"
		encoded, err := encoder.EncodeUserID(userID)
		require.NoError(t, err)

		// Портим hex-строку (меняем последний символ)
		corrupted := encoded[:len(encoded)-1] + "0"
		cookie := &http.Cookie{Value: corrupted}

		_, err = encoder.DecodeUserID(cookie)

		assert.Error(t, err)
	})

	t.Run("fail with wrong secret", func(t *testing.T) {
		encoder1, err := NewCookieEncoder("secret1")
		require.NoError(t, err)
		encoder2, err := NewCookieEncoder("secret2")
		require.NoError(t, err)

		userID := "test-user-id"
		encoded, err := encoder1.EncodeUserID(userID)
		require.NoError(t, err)

		cookie := &http.Cookie{Value: encoded}

		// Пытаемся расшифровать другим энкодером
		_, err = encoder2.DecodeUserID(cookie)

		assert.Error(t, err)
	})

	t.Run("fail on empty cookie value", func(t *testing.T) {
		cookie := &http.Cookie{Value: ""}

		_, err := encoder.DecodeUserID(cookie)

		assert.Error(t, err)
	})

	t.Run("fail on too short encrypted data", func(t *testing.T) {
		// Создаем валидный hex, но слишком короткий для расшифровки
		cookie := &http.Cookie{Value: "0102"}

		_, err := encoder.DecodeUserID(cookie)

		assert.Error(t, err)
	})
}

func TestCookieEncoder_EncodeDecodeUserID_Integration(t *testing.T) {
	encoder, err := NewCookieEncoder("integration-secret")
	require.NoError(t, err)

	testCases := []struct {
		name   string
		userID string
	}{
		{"UUID", uuid.New().String()},
		{"simple string", "user123"},
		{"with special chars", "user@example.com"},
		{"numeric", "12345"},
		{"empty", ""},
		{"long string", string(make([]byte, 1000))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := encoder.EncodeUserID(tc.userID)
			require.NoError(t, err)

			cookie := &http.Cookie{Value: encoded}
			decoded, err := encoder.DecodeUserID(cookie)
			require.NoError(t, err)

			assert.Equal(t, tc.userID, decoded)
		})
	}
}
