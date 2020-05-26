package secretresolver

import (
	"context"
	"fmt"
	"os"
	"strings"
)

const (
	defaultSecretPrefix = "secret://"
)

// Config represents a configuration for Resolve.
type Config struct {
	secretPrefix string
}

// Option represents a function that sets optional values in *Config.
type Option func(c *Config)

// WithSecretPrefix returns Option that sets *Config.secretPrefix with the optional value.
func WithSecretPrefix(p string) Option {
	return func(c *Config) { c.secretPrefix = p }
}

// GetSecretValueFunc represents a function that gets a secret value from the given key.
type GetSecretValueFunc func(ctx context.Context, key string) (value string, err error)

// Resolve transparently resolves environment variables to set secret values.
//
// This function finds the environment variables with the prefix, then gets a secret value with
// the given function GetSecretValueFunc, and finally replace the environment variable value with the secret value.
func Resolve(ctx context.Context, f GetSecretValueFunc, opts ...Option) error {
	c := &Config{
		secretPrefix: defaultSecretPrefix,
	}

	for _, opt := range opts {
		opt(c)
	}

	for _, e := range os.Environ() {
		slugs := strings.SplitN(e, "=", 2)
		if len(slugs) != 2 {
			continue
		}

		envKey, envValue := slugs[0], slugs[1]

		if !strings.HasPrefix(envValue, c.secretPrefix) {
			continue
		}

		secretRef := strings.TrimPrefix(envValue, c.secretPrefix)

		secretVal, err := f(ctx, secretRef)
		if err != nil {
			return fmt.Errorf("failed to resolve %q: %w", envKey, err)
		}

		// Replace secret references in environment variables with secret values.
		os.Setenv(envKey, secretVal)
	}

	return nil
}
