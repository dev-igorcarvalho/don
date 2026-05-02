package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDBConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  DBConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: DBConfig{
				Driver: "postgres",
				DSN:    "postgres://localhost:5432",
			},
			wantErr: false,
		},
		{
			name: "missing driver",
			config: DBConfig{
				DSN: "postgres://localhost:5432",
			},
			wantErr: true,
		},
		{
			name: "missing dsn",
			config: DBConfig{
				Driver: "postgres",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBConfig_ToPkgConfig(t *testing.T) {
	dbConfig := DBConfig{
		Driver:          "postgres",
		DSN:             "postgres://localhost:5432",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
		Warmup:          true,
		ConnectTimeout:  5 * time.Second,
	}

	pkgConfig := dbConfig.ToPkgConfig()

	assert.Equal(t, dbConfig.Driver, pkgConfig.Driver)
	assert.Equal(t, dbConfig.DSN, pkgConfig.DSN)
	assert.Equal(t, dbConfig.MaxOpenConns, pkgConfig.MaxOpenConns)
	assert.Equal(t, dbConfig.MaxIdleConns, pkgConfig.MaxIdleConns)
	assert.Equal(t, dbConfig.ConnMaxLifetime, pkgConfig.ConnMaxLifetime)
	assert.Equal(t, dbConfig.ConnMaxIdleTime, pkgConfig.ConnMaxIdleTime)
	assert.Equal(t, dbConfig.Warmup, pkgConfig.Warmup)
	assert.Equal(t, dbConfig.ConnectTimeout, pkgConfig.ConnectTimeout)
}
