package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithGzip(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	t.Run("compress response when client accepts gzip and content-type is json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test data"))
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := WithGzip(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))

		// Проверяем, что ответ сжат
		reader, err := gzip.NewReader(rr.Body)
		require.NoError(t, err)
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, "test data", string(decompressed))
	})

	t.Run("compress response when content-type is text/plain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("plain text"))
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "text/plain")

		rr := httptest.NewRecorder()

		handler := WithGzip(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))
	})

	t.Run("no compression when client doesn't accept gzip", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test data"))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := WithGzip(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Content-Encoding"))
		assert.Equal(t, "test data", rr.Body.String())
	})

	t.Run("no compression for unsupported content-type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test data"))
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/xml")

		rr := httptest.NewRecorder()

		handler := WithGzip(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Content-Encoding"))
		assert.Equal(t, "test data", rr.Body.String())
	})

	t.Run("decompress gzip request body", func(t *testing.T) {
		originalData := "compressed request data"

		// Сжимаем данные
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		_, err := gzWriter.Write([]byte(originalData))
		require.NoError(t, err)
		require.NoError(t, gzWriter.Close())

		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")

		rr := httptest.NewRecorder()

		handler := WithGzip(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, originalData, rr.Body.String())
	})

	t.Run("fail on invalid gzip request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not gzipped data"))
		req.Header.Set("Content-Encoding", "gzip")

		rr := httptest.NewRecorder()

		handler := WithGzip(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("both compress response and decompress request", func(t *testing.T) {
		originalData := "bidirectional gzip test"

		// Сжимаем запрос
		var reqBuf bytes.Buffer
		gzWriter := gzip.NewWriter(&reqBuf)
		_, err := gzWriter.Write([]byte(originalData))
		require.NoError(t, err)
		require.NoError(t, gzWriter.Close())

		req := httptest.NewRequest(http.MethodPost, "/", &reqBuf)
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := WithGzip(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))

		// Распаковываем ответ
		reader, err := gzip.NewReader(rr.Body)
		require.NoError(t, err)
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, originalData, string(decompressed))
	})

	t.Run("empty request body with gzip encoding", func(t *testing.T) {
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		require.NoError(t, gzWriter.Close())

		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")

		rr := httptest.NewRecorder()

		handler := WithGzip(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Body.String())
	})

	t.Run("large data compression", func(t *testing.T) {
		largeData := strings.Repeat("test data ", 10000)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeData))
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := WithGzip(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))

		// Проверяем, что сжатие эффективно
		assert.Less(t, rr.Body.Len(), len(largeData))

		// Распаковываем и проверяем
		reader, err := gzip.NewReader(rr.Body)
		require.NoError(t, err)
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, largeData, string(decompressed))
	})

	t.Run("content-type with charset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		rr := httptest.NewRecorder()

		handler := WithGzip(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))
	})
}

func TestGzipWriter(t *testing.T) {
	t.Run("write and close", func(t *testing.T) {
		rr := httptest.NewRecorder()
		cw := newCompressWriter(rr)

		data := []byte("test data")
		n, err := cw.Write(data)

		require.NoError(t, err)
		assert.Equal(t, len(data), n)
		require.NoError(t, cw.Close())
	})

	t.Run("write header with success status", func(t *testing.T) {
		rr := httptest.NewRecorder()
		cw := newCompressWriter(rr)

		cw.WriteHeader(http.StatusOK)

		assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))
	})

	t.Run("write header with error status", func(t *testing.T) {
		rr := httptest.NewRecorder()
		cw := newCompressWriter(rr)

		cw.WriteHeader(http.StatusBadRequest)

		assert.Empty(t, rr.Header().Get("Content-Encoding"))
	})
}

func TestCompressReader(t *testing.T) {
	t.Run("read compressed data", func(t *testing.T) {
		originalData := "test data for compression"

		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		_, err := gzWriter.Write([]byte(originalData))
		require.NoError(t, err)
		require.NoError(t, gzWriter.Close())

		cr, err := newCompressReader(nil, io.NopCloser(&buf))
		require.NoError(t, err)
		defer cr.Close()

		result, err := io.ReadAll(cr)
		require.NoError(t, err)
		assert.Equal(t, originalData, string(result))
	})
}
