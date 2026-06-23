package primitives

import (
	"context"
	"errors"
	"testing"
)

func TestPipeline_Run(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var beforeCalled, fnCalled, afterCalled bool
		p := &Pipeline{
			Name: "test-pipeline",
			Before: func(ctx context.Context) error {
				beforeCalled = true
				return nil
			},
			After: func(ctx context.Context) error {
				afterCalled = true
				return nil
			},
			fn: func(ctx context.Context) error {
				fnCalled = true
				return nil
			},
		}

		err := p.Run(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !beforeCalled {
			t.Error("expected Before to be called")
		}
		if !fnCalled {
			t.Error("expected fn to be called")
		}
		if !afterCalled {
			t.Error("expected After to be called")
		}
	})

	t.Run("ContextCancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		p := NewPipeline("test", func(ctx context.Context) error { return nil })
		err := p.Run(ctx)
		if err == nil {
			t.Error("expected error for cancelled context, got nil")
		}
	})

	t.Run("InvalidPipeline", func(t *testing.T) {
		// Empty name
		p1 := &Pipeline{fn: func(ctx context.Context) error { return nil }}
		err := p1.Run(context.Background())
		if err == nil || err.Error() != "pipeline name is required" {
			t.Errorf("expected 'pipeline name is required' error, got %v", err)
		}

		// Nil fn
		p2 := &Pipeline{Name: "test-pipeline"}
		err = p2.Run(context.Background())
		if err == nil || err.Error() != "pipeline fn is required" {
			t.Errorf("expected 'pipeline fn is required' error, got %v", err)
		}
	})

	t.Run("BeforeError", func(t *testing.T) {
		var fnCalled, afterCalled bool
		p := &Pipeline{
			Name: "test",
			Before: func(ctx context.Context) error {
				return errors.New("before error")
			},
			fn: func(ctx context.Context) error {
				fnCalled = true
				return nil
			},
			After: func(ctx context.Context) error {
				afterCalled = true
				return nil
			},
		}

		err := p.Run(context.Background())
		if err == nil {
			t.Fatal("expected error from Before, got nil")
		}

		if fnCalled {
			t.Error("expected fn NOT to be called")
		}
		if afterCalled {
			t.Error("expected After NOT to be called")
		}
	})

	t.Run("FnError", func(t *testing.T) {
		var afterCalled bool
		p := &Pipeline{
			Name: "test",
			fn: func(ctx context.Context) error {
				return errors.New("fn error")
			},
			After: func(ctx context.Context) error {
				afterCalled = true
				return nil
			},
		}

		err := p.Run(context.Background())
		if err == nil {
			t.Fatal("expected error from fn, got nil")
		}

		if !afterCalled {
			t.Error("expected After to be called even if fn fails")
		}
	})

	t.Run("AfterError", func(t *testing.T) {
		p := &Pipeline{
			Name: "test",
			fn:   func(ctx context.Context) error { return nil },
			After: func(ctx context.Context) error {
				return errors.New("after error")
			},
		}

		err := p.Run(context.Background())
		if err == nil {
			t.Fatal("expected error from After, got nil")
		}
	})

	t.Run("BothFnAndAfterError", func(t *testing.T) {
		fnErr := errors.New("fn error")
		afterErr := errors.New("after error")

		p := &Pipeline{
			Name: "test",
			fn: func(ctx context.Context) error {
				return fnErr
			},
			After: func(ctx context.Context) error {
				return afterErr
			},
		}

		err := p.Run(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// Should prioritize fnErr
		if !errors.Is(err, fnErr) && err.Error() != "test: fn error" {
			t.Errorf("expected fn error, got %v", err)
		}
	})

	t.Run("SuccessWithNilAfter", func(t *testing.T) {
		p := &Pipeline{
			Name: "test-nil-after",
			fn: func(ctx context.Context) error {
				return nil
			},
		}

		err := p.Run(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
