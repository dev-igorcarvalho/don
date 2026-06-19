package lifecycle

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockShutdownable struct {
	name     string
	delay    time.Duration
	err      error
	called   bool
	callTime time.Time
	mu       sync.Mutex
}

func (m *mockShutdownable) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	m.called = true
	m.callTime = time.Now()
	m.mu.Unlock()

	select {
	case <-time.After(m.delay):
		return m.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestManager_ShutdownOrder(t *testing.T) {
	m := NewManager(1 * time.Second)

	s1 := &mockShutdownable{name: "first"}
	s2 := &mockShutdownable{name: "second"}

	// Register s1 then s2. Shutdown should be s2 then s1 (LIFO).
	m.Register("s1", s1)
	m.Register("s2", s2)

	err := m.shutdown(context.Background())
	assert.NoError(t, err)

	assert.True(t, s1.called)
	assert.True(t, s2.called)
	assert.True(t, s1.callTime.After(s2.callTime) || s1.callTime.Equal(s2.callTime), "s1 should be called after s2 (LIFO)")
}

func TestManager_Timeout(t *testing.T) {
	// Manager with 50ms timeout
	m := NewManager(50 * time.Millisecond)

	// Component that takes 200ms to shutdown
	s1 := &mockShutdownable{name: "slow", delay: 200 * time.Millisecond}
	m.Register("s1", s1)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := m.shutdown(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), context.DeadlineExceeded.Error())
}

func TestManager_ErrorAggregation(t *testing.T) {
	m := NewManager(1 * time.Second)

	err1 := errors.New("failed 1")
	err2 := errors.New("failed 2")

	s1 := &mockShutdownable{name: "s1", err: err1}
	s2 := &mockShutdownable{name: "s2", err: err2}

	m.Register("s1", s1)
	m.Register("s2", s2)

	err := m.shutdown(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed 1")
	assert.Contains(t, err.Error(), "failed 2")
}

func TestManager_RegisterFunc(t *testing.T) {
	m := NewManager(1 * time.Second)
	called := false
	m.RegisterFunc("fn", func(ctx context.Context) error {
		called = true
		return nil
	})

	err := m.shutdown(context.Background())
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestManager_Wait_ContextCancel(t *testing.T) {
	m := NewManager(1 * time.Second)
	s1 := &mockShutdownable{name: "s1"}
	m.Register("s1", s1)

	ctx, cancel := context.WithCancel(context.Background())

	// Run Wait in a goroutine
	done := make(chan error)
	go func() {
		done <- m.Wait(ctx)
	}()

	// Cancel the context to simulate an external trigger (or signal)
	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
		assert.True(t, s1.called)
	case <-time.After(2 * time.Second):
		t.Fatal("Wait did not return after context cancellation")
	}
}
