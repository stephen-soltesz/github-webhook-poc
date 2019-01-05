// Package githubx extends operations in the google/go-github package.
package githubx

import (
	"testing"
)

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
