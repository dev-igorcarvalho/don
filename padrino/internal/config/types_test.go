package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSqlConfig_ToSqlConnectorConfig(t *testing.T) {
	dbConfig := SqlConfig{
		Driver:          "postgres",
		DSN:             "postgres://localhost:5432",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
		Warmup:          true,
		ConnectTimeout:  5 * time.Second,
	}

	pkgConfig, err := dbConfig.ToSqlConnectorConfig()
	assert.NoError(t, err)
	assert.Equal(t, dbConfig.Driver, pkgConfig.Driver)
	assert.Equal(t, dbConfig.DSN, pkgConfig.DSN)
	assert.Equal(t, dbConfig.MaxOpenConns, pkgConfig.MaxOpenConnections)
	assert.Equal(t, dbConfig.MaxIdleConns, pkgConfig.MaxIdleConnections)
	assert.Equal(t, dbConfig.ConnMaxLifetime, pkgConfig.ConnectionsMaxLifetime)
	assert.Equal(t, dbConfig.ConnMaxIdleTime, pkgConfig.ConnectionsMaxIdleTime)
	assert.Equal(t, dbConfig.Warmup, pkgConfig.Warmup)
	assert.Equal(t, dbConfig.ConnectTimeout, pkgConfig.ConnectTimeout)
}
