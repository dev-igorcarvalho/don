package config

type AppConfig struct {
	Name        string `env:"NAME,required=true"`
	Version     string `env:"VERSION,required=true"`
	Environment string `env:"ENVIRONMENT,required=true"`
	HTTPPort    string `env:"HTTP_PORT,default=8080"`
}
