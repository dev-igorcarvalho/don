// ---
// title: Application Configuration Types
// description: Defines the structures and validation logic for application and database configurations.
// last_updated: 2026-05-03
// type: Configuration
// ---

package config

import (
	"time"

	"github.com/dev-igorcarvalho/don/pkg/database"
)

type AppConfig struct {
	Name        string `env:"NAME,required=true"`
	Version     string `env:"VERSION,required=true"`
	Environment string `env:"ENVIRONMENT,default=SANDBOX"`
	HTTPPort    string `env:"HTTP_PORT,default=8080"`

	RateLimitEnabled bool    `env:"RATE_LIMIT_ENABLED,default=false"`
	RateLimitRPS     float64 `env:"RATE_LIMIT_RPS,default=10"`
	RateLimitBurst   int     `env:"RATE_LIMIT_BURST,default=20"`
}

type SqlDbConfig struct {
	Writer SqlConfig
	Reader SqlConfig
}

type SqlConfig struct {
	Driver          string        `env:"DRIVER,required=true"`
	DSN             string        `env:"DSN,required=true"`
	MaxOpenConns    int           `env:"MAX_OPEN_CONNS,default=25"`
	MaxIdleConns    int           `env:"MAX_IDLE_CONNS,default=25"`
	ConnMaxLifetime time.Duration `env:"CONN_MAX_LIFETIME,default=1h"`
	ConnMaxIdleTime time.Duration `env:"CONN_MAX_IDLE_TIME,default=30m"`
	ConnectTimeout  time.Duration `env:"CONNECT_TIMEOUT,default=5s"`
	QueryTimeout    time.Duration `env:"QUERY_TIMEOUT,default=5s"`
	Warmup          bool          `env:"WARMUP,default=true"`
}

func (c SqlConfig) ToSqlConnectorConfig() (database.Config, error) {
	return database.NewConfig(
		c.Driver,
		c.DSN,
		c.MaxOpenConns,
		c.MaxIdleConns,
		c.ConnMaxLifetime,
		c.ConnMaxIdleTime,
		c.Warmup,
		c.ConnectTimeout,
		c.QueryTimeout,
	)
}
