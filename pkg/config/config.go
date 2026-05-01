// ---
// title: Configuration Loader
// description: Generic configuration loader that unmarshals environment variables into structured Go types.
// last_updated: 2026-05-01
// type: Utility
// ---

// Package config provides a generic configuration loader that unmarshals
// environment variables into structured Go types.
package config

import (
	"fmt"
	"os"

	goenv "github.com/Netflix/go-env"
)

// Override allows forcing environment variable values before loading the configuration.
type Override struct {
	EnvVar string
	Value  string
}

func (o *Override) isValid() bool {
	return o.EnvVar != "" && o.Value != ""
}

// Load populates a configuration struct of type T from environment variables.
// It optionally accepts Overrides to set specific environment variables before unmarshaling.
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
