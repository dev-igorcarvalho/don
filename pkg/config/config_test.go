package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	Name    string `env:"TEST_NAME"`
	Port    int    `env:"TEST_PORT"`
	Enabled bool   `env:"TEST_ENABLED"`
}

func TestLoad(t *testing.T) {
	t.Run("load with overrides", func(t *testing.T) {
		overrides := []Override{
			{EnvVar: "TEST_NAME", Value: "test-app"},
			{EnvVar: "TEST_PORT", Value: "8080"},
			{EnvVar: "TEST_ENABLED", Value: "true"},
		}

		cfg, err := Load[TestConfig](overrides...)
		require.NoError(t, err)

		assert.Equal(t, "test-app", cfg.Name)
		assert.Equal(t, 8080, cfg.Port)
		assert.True(t, cfg.Enabled)
	})

	t.Run("load from existing environment", func(t *testing.T) {
		t.Setenv("TEST_NAME", "env-app")
		t.Setenv("TEST_PORT", "9090")
		t.Setenv("TEST_ENABLED", "false")

		cfg, err := Load[TestConfig]()
		require.NoError(t, err)

		assert.Equal(t, "env-app", cfg.Name)
		assert.Equal(t, 9090, cfg.Port)
		assert.False(t, cfg.Enabled)
	})

	t.Run("invalid overrides should be ignored", func(t *testing.T) {
		t.Setenv("TEST_NAME", "original")
		
		overrides := []Override{
			{EnvVar: "", Value: "ignored"},
			{EnvVar: "TEST_NAME", Value: ""},
		}

		cfg, err := Load[TestConfig](overrides...)
		require.NoError(t, err)

		assert.Equal(t, "original", cfg.Name)
	})

	t.Run("unmarshal error", func(t *testing.T) {
		t.Setenv("TEST_PORT", "not-a-number")

		_, err := Load[TestConfig]()
		assert.Error(t, err)
	})
}

func TestOverride_isValid(t *testing.T) {
	tests := []struct {
		name     string
		override Override
		want     bool
	}{
		{"valid", Override{EnvVar: "KEY", Value: "VAL"}, true},
		{"empty EnvVar", Override{EnvVar: "", Value: "VAL"}, false},
		{"empty Value", Override{EnvVar: "KEY", Value: ""}, false},
		{"both empty", Override{EnvVar: "", Value: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.override.isValid())
		})
	}
}
