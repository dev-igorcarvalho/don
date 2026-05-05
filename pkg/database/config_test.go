package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		cfg, err := NewConfig(
			"postgres",
			"postgres://user:pass@localhost:5432/db",
			10,
			5,
			time.Hour,
			time.Minute,
			true,
			5*time.Second,
		)

		require.NoError(t, err)
		assert.Equal(t, "postgres", cfg.Driver)
		assert.Equal(t, 10, cfg.MaxOpenConnections)
		assert.Equal(t, 5, cfg.MaxIdleConnections)
		assert.True(t, cfg.Warmup)
	})

	t.Run("invalid configuration", func(t *testing.T) {
		_, err := NewConfig("", "", -1, -1, -1, -1, false, -1)
		assert.Error(t, err)
	})
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "valid config",
			config: Config{
				Driver:                 "postgres",
				DSN:                    "dsn",
				MaxOpenConnections:     10,
				MaxIdleConnections:     5,
				ConnectionsMaxLifetime: time.Hour,
				ConnectionsMaxIdleTime: time.Minute,
				ConnectTimeout:         time.Second,
			},
			wantErr: nil,
		},
		{
			name: "missing driver",
			config: Config{
				DSN: "dsn",
			},
			wantErr: ErrInvalidDriver,
		},
		{
			name: "missing dsn",
			config: Config{
				Driver: "postgres",
			},
			wantErr: ErrInvalidDSN,
		},
		{
			name: "negative max open connections",
			config: Config{
				Driver:             "postgres",
				DSN:                "dsn",
				MaxOpenConnections: -1,
			},
			wantErr: ErrInvalidMaxOpenConns,
		},
		{
			name: "negative max idle connections",
			config: Config{
				Driver:             "postgres",
				DSN:                "dsn",
				MaxOpenConnections: 10,
				MaxIdleConnections: -1,
			},
			wantErr: ErrInvalidMaxIdleConns,
		},
		{
			name: "negative connections max lifetime",
			config: Config{
				Driver:                 "postgres",
				DSN:                    "dsn",
				MaxOpenConnections:     10,
				MaxIdleConnections:     5,
				ConnectionsMaxLifetime: -1,
			},
			wantErr: ErrInvalidConnMaxLifetime,
		},
		{
			name: "negative connections max idle time",
			config: Config{
				Driver:                 "postgres",
				DSN:                    "dsn",
				MaxOpenConnections:     10,
				MaxIdleConnections:     5,
				ConnectionsMaxLifetime: time.Hour,
				ConnectionsMaxIdleTime: -1,
			},
			wantErr: ErrInvalidConnMaxIdleTime,
		},
		{
			name: "negative connect timeout",
			config: Config{
				Driver:                 "postgres",
				DSN:                    "dsn",
				MaxOpenConnections:     10,
				MaxIdleConnections:     5,
				ConnectionsMaxLifetime: time.Hour,
				ConnectionsMaxIdleTime: time.Minute,
				ConnectTimeout:         -1,
			},
			wantErr: ErrInvalidConnectTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
