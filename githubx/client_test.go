// Package githubx extends operations in the google/go-github package.
package githubx

import (
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name             string
		installationID   int64
		githubAuthToken  string
		githubPrivateKey string
		githubAppID      string
		wantNil          bool
	}{
		{
			name:            "success",
			installationID:  0,
			githubAuthToken: "test",
		},
		{
			name:             "success",
			installationID:   1,
			githubPrivateKey: "testdata/unused_insecure_rsa_key.pem",
			githubAppID:      "1",
		},
		{
			name:           "missing-auth-token",
			installationID: 0,
			wantNil:        true,
		},
		{
			name:           "missing-private-key",
			installationID: 1,
			wantNil:        true,
		},
		{
			name:             "missing-app-id",
			installationID:   1,
			githubPrivateKey: "testdata/unused_insecure_rsa_key.pem",
			wantNil:          true,
		},
		{
			name:             "invalid-app-id",
			installationID:   1,
			githubPrivateKey: "testdata/unused_insecure_rsa_key.pem",
			githubAppID:      "NOT-A-NUMBER",
			wantNil:          true,
		},
	}
	for _, tt := range tests {
		if tt.githubAuthToken != "" {
			os.Setenv("GITHUB_AUTH_TOKEN", tt.githubAuthToken)
		} else {
			os.Unsetenv("GITHUB_AUTH_TOKEN")
		}
		if tt.githubPrivateKey != "" {
			os.Setenv("GITHUB_PRIVATE_KEY", tt.githubPrivateKey)
		} else {
			os.Unsetenv("GITHUB_PRIVATE_KEY")
		}
		if tt.githubAppID != "" {
			os.Setenv("GITHUB_APP_ID", tt.githubAppID)
		} else {
			os.Unsetenv("GITHUB_APP_ID")
		}

		t.Run(tt.name, func(t *testing.T) {
			got := NewClient(tt.installationID)
			if got == nil && !tt.wantNil {
				t.Errorf("NewClient() = %v, want nil", got)
			}
		})
	}
}

func TestNewPersonalClient(t *testing.T) {
	_ = NewPersonalClient("anything")
}

func TestNewAppClient(t *testing.T) {
	tests := []struct {
		name       string
		privateKey string
		wantNil    bool
	}{
		{
			name:       "error returns nil",
			privateKey: "",
			wantNil:    true,
		},
		{
			name:       "success",
			privateKey: "testdata/unused_insecure_rsa_key.pem",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAppClient(tt.privateKey, 1, 1)
			if got == nil && !tt.wantNil {
				t.Errorf("NewAppClient() = %v, want nil", got)
			}
		})
	}
}
