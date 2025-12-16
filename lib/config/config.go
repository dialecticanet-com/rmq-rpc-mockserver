// Package config provides a way to load configuration from environment variables and environment file.
package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Load loads the configuration from the environment variables.
func Load(config any, opts ...Option) error {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	// Load config from file
	if o.envFile != "" {
		if err := godotenv.Load(o.envFile); err != nil {
			return fmt.Errorf("error is occurred on '%s' environment file: %w", o.envFile, err)
		}
	}

	//  Populate the config struct with the environment variables
	if err := envconfig.Process(o.envPrefix, config); err != nil {
		return fmt.Errorf("error is occurred on loading config: %w", err)
	}

	return nil
}
