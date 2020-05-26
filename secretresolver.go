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

type SecretResolver interface {
	Resolve(context.Context) error
}

type secretResolver struct {
	secretManager SecretManager

	secretPrefix string
}

// Guarantee *resolver implements Resolver.
var _ SecretResolver = (*secretResolver)(nil)

type Option func(r *secretResolver)

func WithSecretPrefix(p string) Option {
	return func(r *secretResolver) { r.secretPrefix = p }
}

func New(secretManager SecretManager, opts ...Option) SecretResolver {
	r := &secretResolver{
		secretManager: secretManager,
		secretPrefix:  defaultSecretPrefix,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *secretResolver) Resolve(ctx context.Context) error {
	for _, e := range os.Environ() {
		slugs := strings.SplitN(e, "=", 2)
		if len(slugs) != 2 {
			continue
		}

		envKey, envValue := slugs[0], slugs[1]

		if !strings.HasPrefix(envValue, r.secretPrefix) {
			continue
		}

		secretRef := strings.TrimPrefix(envValue, r.secretPrefix)

		secretVal, err := r.secretManager.GetSecretValue(ctx, secretRef)
		if err != nil {
			return fmt.Errorf("failed to resolve %q: %w", envKey, err)
		}

		// Replace secret references in environment variables with secret values.
		os.Setenv(envKey, secretVal)
	}

	return nil
}
