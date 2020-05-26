package secretresolver

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_secretResolver_Resolve(t *testing.T) {
	tests := []struct {
		name        string
		envKey      string
		envValue    string
		secretValue string
		want        string
	}{
		{
			name:     "keep non-secret env",
			envKey:   "KEY1",
			envValue: "VALUE1",
			want:     "VALUE1",
		},
		{
			name:        "resolve secret",
			envKey:      "KEY2",
			envValue:    "secret://projects/my-project/secrets/my-secret/versions/123",
			secretValue: "VALUE2",
			want:        "VALUE2",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			func() {
				setenv(t, tt.envKey, tt.envValue)

				sm := newFakeManager()

				if tt.secretValue != "" {
					sm.SetSecretValue(strings.TrimPrefix(tt.envValue, defaultSecretPrefix), tt.secretValue) // Set fake secret value.
				}

				r := New(sm)

				if err := r.Resolve(context.Background()); err != nil {
					t.Fatalf("err: %v", err)
				}

				got := os.Getenv(tt.envKey)
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Fatalf("(-want +got):\n%s", diff)
				}
			}()
		})
	}
}

func setenv(t *testing.T, key, value string) {
	t.Helper()

	prev := os.Getenv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Setenv(key, prev); err != nil {
			t.Fatal(err)
		}
	})
}
