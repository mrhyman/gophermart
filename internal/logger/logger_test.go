package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestWithinContext(t *testing.T) {
	t.Run("add logger to context", func(t *testing.T) {
		ctx := context.Background()
		logger := New()

		newCtx := WithinContext(ctx, logger)

		assert.NotEqual(t, ctx, newCtx)
		assert.NotNil(t, newCtx.Value(ctxKey{}))
	})

	t.Run("replace logger in context", func(t *testing.T) {
		logger1 := New()
		logger2 := New()
		ctx := WithinContext(context.Background(), logger1)

		newCtx := WithinContext(ctx, logger2)

		retrievedLogger := FromContext(newCtx)
		assert.Equal(t, logger2, retrievedLogger)
	})
}

func TestFromContext(t *testing.T) {
	t.Run("get logger from context", func(t *testing.T) {
		logger := New()
		ctx := WithinContext(context.Background(), logger)

		retrieved := FromContext(ctx)

		assert.Equal(t, logger, retrieved)
	})

	t.Run("return default logger when not in context", func(t *testing.T) {
		ctx := context.Background()

		logger := FromContext(ctx)

		assert.NotNil(t, logger)
	})

	t.Run("return default logger when nil logger in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ctxKey{}, nil)

		logger := FromContext(ctx)

		assert.NotNil(t, logger)
	})

	t.Run("return default logger when wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ctxKey{}, "not a logger")

		logger := FromContext(ctx)

		assert.NotNil(t, logger)
	})
}

func TestNew(t *testing.T) {
	t.Run("create new logger", func(t *testing.T) {
		logger := New()

		assert.NotNil(t, logger)
	})

	t.Run("different instances on multiple calls", func(t *testing.T) {
		logger1 := New()
		logger2 := New()

		assert.NotEqual(t, logger1, logger2)
	})
}

func TestNewWithCore(t *testing.T) {
	t.Run("create logger with custom core", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)

		logger := NewWithCore(core)

		require.NotNil(t, logger)
		logger.Info("test message")
	})

	t.Run("logger uses provided core", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := NewWithCore(core)

		logger.Info("test info")
		logger.Debug("test debug") // не должно записаться

		assert.Equal(t, 1, recorded.Len())
		assert.Equal(t, "test info", recorded.All()[0].Message)
	})
}

func TestContextIntegration(t *testing.T) {
	t.Run("full workflow", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := NewWithCore(core)

		ctx := context.Background()
		ctx = WithinContext(ctx, logger)

		retrievedLogger := FromContext(ctx)
		retrievedLogger.Info("integration test")

		assert.Equal(t, 1, recorded.Len())
		assert.Equal(t, "integration test", recorded.All()[0].Message)
	})

	t.Run("nested contexts", func(t *testing.T) {
		core1, recorded1 := observer.New(zapcore.InfoLevel)
		core2, recorded2 := observer.New(zapcore.InfoLevel)
		logger1 := NewWithCore(core1)
		logger2 := NewWithCore(core2)

		ctx := context.Background()
		ctx1 := WithinContext(ctx, logger1)
		ctx2 := WithinContext(ctx1, logger2)

		FromContext(ctx1).Info("from ctx1")
		FromContext(ctx2).Info("from ctx2")

		assert.Equal(t, 1, recorded1.Len())
		assert.Equal(t, 1, recorded2.Len())
		assert.Equal(t, "from ctx1", recorded1.All()[0].Message)
		assert.Equal(t, "from ctx2", recorded2.All()[0].Message)
	})
}
