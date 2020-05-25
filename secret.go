package secretresolver

import (
	"context"
	"sync"
)

// SecretManager represents a secret manager that accesses secrets.
type SecretManager interface {
	GetSecretValue(ctx context.Context, name string) (string, error)
}

// fakeSecretManager represents a fake secret manager for testing.
type fakeSecretManager struct {
	values map[string]string

	mu sync.RWMutex
}

// Guarantee *FakeManager implements SecretManager.
var _ SecretManager = (*fakeSecretManager)(nil)

func newFakeManager() *fakeSecretManager {
	return &fakeSecretManager{
		values: make(map[string]string),
		mu:     sync.RWMutex{},
	}
}

func (m *fakeSecretManager) GetSecretValue(ctx context.Context, name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.values[name], nil
}

func (m *fakeSecretManager) SetSecretValue(key, value string) {
	m.mu.Lock()
	m.values[key] = value
	m.mu.Unlock()
}
