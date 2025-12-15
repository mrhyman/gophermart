package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCrypto(t *testing.T) {
	t.Run("success with valid secret", func(t *testing.T) {
		crypto, err := NewCrypto("my-secret-key")

		require.NoError(t, err)
		assert.NotNil(t, crypto)
		assert.NotNil(t, crypto.aesgcm)
	})

	t.Run("success with empty secret", func(t *testing.T) {
		crypto, err := NewCrypto("")

		require.NoError(t, err)
		assert.NotNil(t, crypto)
	})

	t.Run("different secrets create different instances", func(t *testing.T) {
		crypto1, err1 := NewCrypto("secret1")
		crypto2, err2 := NewCrypto("secret2")

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, crypto1, crypto2)
	})
}

func TestCrypto_Encrypt(t *testing.T) {
	crypto, err := NewCrypto("test-secret")
	require.NoError(t, err)

	t.Run("encrypt plain text", func(t *testing.T) {
		plaintext := []byte("Hello, World!")

		ciphertext, err := crypto.Encrypt(plaintext)

		require.NoError(t, err)
		assert.NotEmpty(t, ciphertext)
		assert.NotEqual(t, plaintext, ciphertext)
		assert.Greater(t, len(ciphertext), len(plaintext))
	})

	t.Run("encrypt empty data", func(t *testing.T) {
		ciphertext, err := crypto.Encrypt([]byte{})

		require.NoError(t, err)
		assert.NotEmpty(t, ciphertext) // nonce всегда присутствует
	})

	t.Run("different ciphertexts for same plaintext", func(t *testing.T) {
		plaintext := []byte("same data")

		ciphertext1, err1 := crypto.Encrypt(plaintext)
		ciphertext2, err2 := crypto.Encrypt(plaintext)

		require.NoError(t, err1)
		require.NoError(t, err2)
		// Из-за разных nonce шифртексты должны отличаться
		assert.NotEqual(t, ciphertext1, ciphertext2)
	})

	t.Run("encrypt large data", func(t *testing.T) {
		plaintext := make([]byte, 10000)
		for i := range plaintext {
			plaintext[i] = byte(i % 256)
		}

		ciphertext, err := crypto.Encrypt(plaintext)

		require.NoError(t, err)
		assert.Greater(t, len(ciphertext), len(plaintext))
	})
}

func TestCrypto_Decrypt(t *testing.T) {
	crypto, err := NewCrypto("test-secret")
	require.NoError(t, err)

	t.Run("decrypt valid ciphertext", func(t *testing.T) {
		plaintext := []byte("Secret Message")
		ciphertext, err := crypto.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := crypto.Decrypt(ciphertext)

		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("decrypt empty encrypted data", func(t *testing.T) {
		ciphertext, err := crypto.Encrypt([]byte{})
		require.NoError(t, err)

		decrypted, err := crypto.Decrypt(ciphertext)

		require.NoError(t, err)
		assert.Len(t, decrypted, 0)
	})

	t.Run("fail on too short ciphertext", func(t *testing.T) {
		shortCiphertext := []byte{1, 2, 3}

		_, err := crypto.Decrypt(shortCiphertext)

		assert.Error(t, err)
		assert.Equal(t, "ciphertext too short", err.Error())
	})

	t.Run("fail on corrupted ciphertext", func(t *testing.T) {
		plaintext := []byte("data")
		ciphertext, err := crypto.Encrypt(plaintext)
		require.NoError(t, err)

		// Портим данные
		ciphertext[len(ciphertext)-1] ^= 0xFF

		_, err = crypto.Decrypt(ciphertext)

		assert.Error(t, err)
	})

	t.Run("fail with wrong secret key", func(t *testing.T) {
		crypto1, err := NewCrypto("secret1")
		require.NoError(t, err)
		crypto2, err := NewCrypto("secret2")
		require.NoError(t, err)

		plaintext := []byte("encrypted with secret1")
		ciphertext, err := crypto1.Encrypt(plaintext)
		require.NoError(t, err)

		// Пытаемся расшифровать другим ключом
		_, err = crypto2.Decrypt(ciphertext)

		assert.Error(t, err)
	})
}

func TestCrypto_EncryptDecrypt_Integration(t *testing.T) {
	crypto, err := NewCrypto("integration-test-secret")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"simple text", []byte("Hello, World!")},
		{"empty", nil},
		{"with special chars", []byte("Test\n\t\r!@#$%^&*()")},
		{"unicode", []byte("Привет, мир! 你好世界")},
		{"binary data", []byte{0x00, 0xFF, 0x01, 0xFE, 0x02, 0xFD}},
		{"large text", []byte(string(make([]byte, 5000)))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := crypto.Encrypt(tc.plaintext)
			require.NoError(t, err)

			decrypted, err := crypto.Decrypt(ciphertext)
			require.NoError(t, err)

			assert.Equal(t, tc.plaintext, decrypted)
		})
	}
}
