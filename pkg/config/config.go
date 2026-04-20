package config

import (
	"fmt"
	"os"

	goenv "github.com/Netflix/go-env"
)

type Override struct {
	EnvVar string
	Value  string
}

func (o *Override) isValid() bool {
	return o.EnvVar != "" && o.Value != ""
}

func Load[T any](overrides ...Override) (*T, error) {
	for _, o := range overrides {
		if o.isValid() {
			err := os.Setenv(o.EnvVar, o.Value)
			if err != nil {
				return nil, err
			}
		}
	}
	var cfg T
	_, err := goenv.UnmarshalFromEnviron(&cfg)
	if err != nil {
		return nil, fmt.Errorf("config: %T failed to load - %w", cfg, err)
	}
	return &cfg, nil
}
