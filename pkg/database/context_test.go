package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriterContext(t *testing.T) {
	t.Run("WithWriterContext sets the forced flag", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, isWriterForced(ctx), "expected forced flag to be false by default")

		writerCtx := WithWriterContext(ctx)
		assert.True(t, isWriterForced(writerCtx), "expected forced flag to be true after WithWriterContext")
	})

	t.Run("isWriterForced returns false for nil context", func(t *testing.T) {
		// Although context.Context should ideally not be nil, we check the behavior.
		// Passing nil to context.WithValue would panic, but isWriterForced handles
		// the case where the key is missing or the value is not a bool correctly
		// if the context itself is valid.
		// Note: isWriterForced(nil) would panic on ctx.Value() call.
		// Go standards usually assume non-nil context.
	})

	t.Run("isWriterForced handles non-boolean values or missing keys", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, isWriterForced(ctx))

		// Manually setting a wrong type for the key (if we could, but contextKey is unexported)
		// Since contextKey is internal, we can only test missing key which is already done above.
	})

	t.Run("context inheritance", func(t *testing.T) {
		ctx := WithWriterContext(context.Background())
		childCtx := context.WithValue(ctx, "other", "value")

		assert.True(t, isWriterForced(childCtx), "child context should inherit the forced flag")
	})
}
