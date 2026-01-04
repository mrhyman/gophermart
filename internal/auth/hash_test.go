package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		password := "mySecretPassword123"

		hash, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)
		assert.Greater(t, len(hash), 50) // bcrypt хеш обычно >50 символов
	})

	t.Run("empty password", func(t *testing.T) {
		hash, err := HashPassword("")

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("different hashes for same password", func(t *testing.T) {
		password := "samePassword"

		hash1, err1 := HashPassword(password)
		hash2, err2 := HashPassword(password)

		require.NoError(t, err1)
		require.NoError(t, err2)
		// bcrypt генерирует разные хеши из-за соли
		assert.NotEqual(t, hash1, hash2)
	})
}

func TestCheckPassword(t *testing.T) {
	t.Run("correct password", func(t *testing.T) {
		password := "correctPassword"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		err = CheckPassword(password, hash)

		assert.NoError(t, err)
	})

	t.Run("incorrect password", func(t *testing.T) {
		password := "correctPassword"
		wrongPassword := "wrongPassword"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		err = CheckPassword(wrongPassword, hash)

		assert.Error(t, err)
	})

	t.Run("empty password against hash", func(t *testing.T) {
		password := "somePassword"
		hash, err := HashPassword(password)
		require.NoError(t, err)

		err = CheckPassword("", hash)

		assert.Error(t, err)
	})

	t.Run("invalid hash format", func(t *testing.T) {
		err := CheckPassword("password", "invalid-hash")

		assert.Error(t, err)
	})

	t.Run("empty hash", func(t *testing.T) {
		err := CheckPassword("password", "")

		assert.Error(t, err)
	})
}
