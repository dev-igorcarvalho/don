package must

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		val := Get(42, nil)
		assert.Equal(t, 42, val)
	})

	t.Run("fatal", func(t *testing.T) {
		if os.Getenv("BE_CRASHY") == "1" {
			Get(0, errors.New("test error"))
			return
		}
		cmd := exec.Command(os.Args[0], "-test.run=TestGet/fatal")
		cmd.Env = append(os.Environ(), "BE_CRASHY=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
		t.Fatalf("process ran with err %v, want exit status 1", err)
	})
}

func TestGetf(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		val := Getf(42, nil, "context %d", 1)
		assert.Equal(t, 42, val)
	})

	t.Run("fatal", func(t *testing.T) {
		if os.Getenv("BE_CRASHY") == "1" {
			Getf(0, errors.New("test error"), "context %d", 1)
			return
		}
		cmd := exec.Command(os.Args[0], "-test.run=TestGetf/fatal")
		cmd.Env = append(os.Environ(), "BE_CRASHY=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
		t.Fatalf("process ran with err %v, want exit status 1", err)
	})
}

func TestSucceed(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		Succeed(nil)
	})

	t.Run("fatal", func(t *testing.T) {
		if os.Getenv("BE_CRASHY") == "1" {
			Succeed(errors.New("test error"))
			return
		}
		cmd := exec.Command(os.Args[0], "-test.run=TestSucceed/fatal")
		cmd.Env = append(os.Environ(), "BE_CRASHY=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
		t.Fatalf("process ran with err %v, want exit status 1", err)
	})
}

func TestSucceedf(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		Succeedf(nil, "context")
	})

	t.Run("fatal", func(t *testing.T) {
		if os.Getenv("BE_CRASHY") == "1" {
			Succeedf(errors.New("test error"), "context %d", 1)
			return
		}
		cmd := exec.Command(os.Args[0], "-test.run=TestSucceedf/fatal")
		cmd.Env = append(os.Environ(), "BE_CRASHY=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
		t.Fatalf("process ran with err %v, want exit status 1", err)
	})
}

func TestBeTrue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		BeTrue(true, "must be true")
	})

	t.Run("fatal", func(t *testing.T) {
		if os.Getenv("BE_CRASHY") == "1" {
			BeTrue(false, "should fail")
			return
		}
		cmd := exec.Command(os.Args[0], "-test.run=TestBeTrue/fatal")
		cmd.Env = append(os.Environ(), "BE_CRASHY=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
		t.Fatalf("process ran with err %v, want exit status 1", err)
	})
}

func TestBeTruef(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		BeTruef(true, "must be %s", "true")
	})

	t.Run("fatal", func(t *testing.T) {
		if os.Getenv("BE_CRASHY") == "1" {
			BeTruef(false, "should fail %d", 1)
			return
		}
		cmd := exec.Command(os.Args[0], "-test.run=TestBeTruef/fatal")
		cmd.Env = append(os.Environ(), "BE_CRASHY=1")
		err := cmd.Run()
		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			return
		}
		t.Fatalf("process ran with err %v, want exit status 1", err)
	})
}

func TestExample(t *testing.T) {
	// Simple example test
	val := Get(10, nil)
	assert.Equal(t, 10, val)
}
