package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		driver := "postgres"
		dsn := "postgres://user:pass@localhost:5432/db"
		maxOpen := 10
		maxIdle := 5
		lifetime := time.Hour
		idleTime := time.Minute
		warmup := true
		timeout := 5 * time.Second
		queryTimeout := 3 * time.Second

		cfg, err := NewConfig(
			driver,
			dsn,
			maxOpen,
			maxIdle,
			lifetime,
			idleTime,
			warmup,
			timeout,
			queryTimeout,
		)

		require.NoError(t, err)
		assert.Equal(t, driver, cfg.Driver)
		assert.Equal(t, dsn, cfg.DSN)
		assert.Equal(t, maxOpen, cfg.MaxOpenConnections)
		assert.Equal(t, maxIdle, cfg.MaxIdleConnections)
		assert.Equal(t, lifetime, cfg.ConnectionsMaxLifetime)
		assert.Equal(t, idleTime, cfg.ConnectionsMaxIdleTime)
		assert.Equal(t, warmup, cfg.Warmup)
		assert.Equal(t, timeout, cfg.ConnectTimeout)
		assert.Equal(t, queryTimeout, cfg.QueryTimeout)
	})

	t.Run("invalid configuration - missing driver", func(t *testing.T) {
		_, err := NewConfig("", "dsn", 1, 1, time.Second, time.Second, false, time.Second, time.Second)
		assert.ErrorIs(t, err, ErrInvalidDriver)
	})

	t.Run("invalid configuration - negative query timeout", func(t *testing.T) {
		_, err := NewConfig("postgres", "dsn", 1, 1, time.Second, time.Second, false, time.Second, -1)
		assert.ErrorIs(t, err, ErrInvalidQueryTimeout)
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
				QueryTimeout:           time.Second,
			},
			wantErr: nil,
		},
		{
			name: "valid config with zero values",
			config: Config{
				Driver:                 "postgres",
				DSN:                    "dsn",
				MaxOpenConnections:     0,
				MaxIdleConnections:     0,
				ConnectionsMaxLifetime: 0,
				ConnectionsMaxIdleTime: 0,
				ConnectTimeout:         0,
				QueryTimeout:           0,
			},
			wantErr: nil,
		},
		{
			name: "missing driver",
			config: Config{
				Driver: "",
				DSN:    "dsn",
			},
			wantErr: ErrInvalidDriver,
		},
		{
			name: "missing dsn",
			config: Config{
				Driver: "postgres",
				DSN:    "",
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
				MaxIdleConnections: -1,
			},
			wantErr: ErrInvalidMaxIdleConns,
		},
		{
			name: "negative connections max lifetime",
			config: Config{
				Driver:                 "postgres",
				DSN:                    "dsn",
				ConnectionsMaxLifetime: -1,
			},
			wantErr: ErrInvalidConnMaxLifetime,
		},
		{
			name: "negative connections max idle time",
			config: Config{
				Driver:                 "postgres",
				DSN:                    "dsn",
				ConnectionsMaxIdleTime: -1,
			},
			wantErr: ErrInvalidConnMaxIdleTime,
		},
		{
			name: "negative connect timeout",
			config: Config{
				Driver:         "postgres",
				DSN:            "dsn",
				ConnectTimeout: -1,
			},
			wantErr: ErrInvalidConnectTimeout,
		},
		{
			name: "negative query timeout",
			config: Config{
				Driver:       "postgres",
				DSN:          "dsn",
				QueryTimeout: -1,
			},
			wantErr: ErrInvalidQueryTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
