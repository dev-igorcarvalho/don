package config

import "time"

type AppConfig struct {
	Name        string `env:"NAME,required=true"`
	Version     string `env:"VERSION,required=true"`
	Environment string `env:"ENVIRONMENT,default=SANDBOX"`
	HTTPPort    string `env:"HTTP_PORT,default=8080"`

	RateLimitEnabled bool    `env:"RATE_LIMIT_ENABLED,default=false"`
	RateLimitRPS     float64 `env:"RATE_LIMIT_RPS,default=10"`
	RateLimitBurst   int     `env:"RATE_LIMIT_BURST,default=20"`

	DB DBConfig
}

type DBConfig struct {
	Driver string `env:"DB_DRIVER,default=postgres"`
	DSN    string `env:"DB_DSN,required=true"`

	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS,default=25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS,default=25"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME,default=1h"`
	ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME,default=30m"`

	ConnectTimeout time.Duration `env:"DB_CONNECT_TIMEOUT,default=5s"`
	QueryTimeout   time.Duration `env:"DB_QUERY_TIMEOUT,default=15s"`

	Warmup bool `env:"DB_WARMUP,default=true"`
}
