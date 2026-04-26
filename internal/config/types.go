package config

type AppConfig struct {
	Name        string `env:"NAME,required=true"`
	Version     string `env:"VERSION,required=true"`
	Environment string `env:"ENVIRONMENT,default=SANDBOX"`
	HTTPPort    string `env:"HTTP_PORT,default=8080"`

	RateLimitEnabled bool    `env:"RATE_LIMIT_ENABLED,default=false"`
	RateLimitRPS     float64 `env:"RATE_LIMIT_RPS,default=10"`
	RateLimitBurst   int     `env:"RATE_LIMIT_BURST,default=20"`
}
