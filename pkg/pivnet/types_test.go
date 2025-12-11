// ABOUTME: Unit tests for Pivnet client types and configuration.
// ABOUTME: Validates configuration parsing and validation logic.
package pivnet

import (
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config with token",
			config: Config{
				Token:           "test-token",
				ProductSlug:     "cf",
				OldVersion:      "6.0.22+LTS-T",
				NewVersion:      "10.2.5+LTS-T",
				NonInteractive:  false,
			},
			wantErr: false,
		},
		{
			name: "missing token",
			config: Config{
				ProductSlug: "cf",
				OldVersion:  "6.0.22",
				NewVersion:  "10.2.5",
			},
			wantErr: true,
		},
		{
			name: "missing product slug",
			config: Config{
				Token:      "test-token",
				OldVersion: "6.0.22",
				NewVersion: "10.2.5",
			},
			wantErr: true,
		},
		{
			name: "missing versions",
			config: Config{
				Token:       "test-token",
				ProductSlug: "cf",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDownloadOptionsValidation(t *testing.T) {
	config := Config{
		Token:       "test-token",
		ProductSlug: "cf",
		OldVersion:  "6.0.22+LTS-T",
		NewVersion:  "10.2.5+LTS-T",
		CacheDir:    "/tmp/tile-diff-cache",
	}

	opts := config.DownloadOptions()

	if opts.ProductSlug != "cf" {
		t.Errorf("Expected ProductSlug 'cf', got '%s'", opts.ProductSlug)
	}
	if opts.CacheDir != "/tmp/tile-diff-cache" {
		t.Errorf("Expected CacheDir '/tmp/tile-diff-cache', got '%s'", opts.CacheDir)
	}
}
