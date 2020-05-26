package secretresolver

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResolve(t *testing.T) {
	type args struct {
		f    GetSecretValueFunc
		opts []Option
	}

	tests := []struct {
		name     string
		args     args
		fakeEnvs map[string]string
		want     map[string]string
	}{
		{
			name: "keep non-secret env",
			fakeEnvs: map[string]string{
				"KEY1": "VALUE1",
			},
			want: map[string]string{
				"KEY1": "VALUE1",
			},
		},
		{
			name: "resolve secret env",
			args: args{
				f: func(_ context.Context, key string) (string, error) {
					s := map[string]string{
						"projects/my-project/secrets/my-secret/versions/123": "VALUE2",
					}
					return s[key], nil
				},
			},
			fakeEnvs: map[string]string{
				"KEY2": "secret://projects/my-project/secrets/my-secret/versions/123",
			},
			want: map[string]string{
				"KEY2": "VALUE2",
			},
		},
		{
			name: "resolve secret env with custom prefix",
			args: args{
				f: func(_ context.Context, key string) (string, error) {
					s := map[string]string{
						"projects/my-project/secrets/my-secret/versions/456": "VALUE3",
					}
					return s[key], nil
				},
				opts: []Option{WithSecretPrefix("s://")},
			},
			fakeEnvs: map[string]string{
				"KEY3": "s://projects/my-project/secrets/my-secret/versions/456",
			},
			want: map[string]string{
				"KEY3": "VALUE3",
			},
		},
		{
			name: "resolve some secret envs and keep some non-secret envs",
			args: args{
				f: func(_ context.Context, key string) (string, error) {
					s := map[string]string{
						"projects/my-project/secrets/my-secret/versions/789": "VALUE4",
					}
					return s[key], nil
				},
			},
			fakeEnvs: map[string]string{
				"KEY4": "secret://projects/my-project/secrets/my-secret/versions/789",
				"KEY5": "VALUE5",
			},
			want: map[string]string{
				"KEY4": "VALUE4",
				"KEY5": "VALUE5",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			func() {
				for k, v := range tt.fakeEnvs {
					setenv(t, k, v)
				}

				if err := Resolve(context.Background(), tt.args.f, tt.args.opts...); err != nil {
					t.Fatalf("err: %v", err)
				}

				for k, v := range tt.want {
					got := os.Getenv(k)
					if diff := cmp.Diff(v, got); diff != "" {
						t.Fatalf("(-want +got):\n%s", diff)
					}
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
