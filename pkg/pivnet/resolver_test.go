// ABOUTME: Unit tests for version resolution and fuzzy matching.
// ABOUTME: Validates exact and fuzzy version matching logic.
package pivnet

import (
	"testing"
)

func TestResolveVersion(t *testing.T) {
	releases := []Release{
		{ID: 1, Version: "6.0.22+LTS-T"},
		{ID: 2, Version: "6.0.21+LTS-T"},
		{ID: 3, Version: "6.0.20+LTS-T"},
		{ID: 4, Version: "10.2.5+LTS-T"},
		{ID: 5, Version: "10.2.4+LTS-T"},
	}

	tests := []struct {
		name           string
		version        string
		nonInteractive bool
		wantVersion    string
		wantMultiple   bool
		wantError      bool
	}{
		{
			name:           "exact match",
			version:        "6.0.22+LTS-T",
			nonInteractive: false,
			wantVersion:    "6.0.22+LTS-T",
			wantMultiple:   false,
			wantError:      false,
		},
		{
			name:           "fuzzy match single result",
			version:        "10.2.5",
			nonInteractive: false,
			wantVersion:    "10.2.5+LTS-T",
			wantMultiple:   false,
			wantError:      false,
		},
		{
			name:           "fuzzy match multiple results - interactive ok",
			version:        "6.0",
			nonInteractive: false,
			wantVersion:    "",
			wantMultiple:   true,
			wantError:      false,
		},
		{
			name:           "fuzzy match multiple results - non-interactive error",
			version:        "6.0",
			nonInteractive: true,
			wantVersion:    "",
			wantMultiple:   false,
			wantError:      true,
		},
		{
			name:           "no match",
			version:        "99.0.0",
			nonInteractive: false,
			wantVersion:    "",
			wantMultiple:   false,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(releases, tt.nonInteractive)
			result, err := resolver.Resolve(tt.version)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.wantMultiple {
				if len(result.Matches) < 2 {
					t.Errorf("Expected multiple matches, got %d", len(result.Matches))
				}
			} else if tt.wantVersion != "" {
				if result.Selected == nil {
					t.Error("Expected selected release, got nil")
					return
				}
				if result.Selected.Version != tt.wantVersion {
					t.Errorf("Expected version %s, got %s", tt.wantVersion, result.Selected.Version)
				}
			}
		})
	}
}

func TestMatchesVersion(t *testing.T) {
	tests := []struct {
		name         string
		fullVersion  string
		searchString string
		wantMatch    bool
	}{
		{
			name:         "exact match",
			fullVersion:  "6.0.22+LTS-T",
			searchString: "6.0.22+LTS-T",
			wantMatch:    true,
		},
		{
			name:         "major.minor match",
			fullVersion:  "6.0.22+LTS-T",
			searchString: "6.0",
			wantMatch:    true,
		},
		{
			name:         "major.minor.patch match",
			fullVersion:  "6.0.22+LTS-T",
			searchString: "6.0.22",
			wantMatch:    true,
		},
		{
			name:         "no match",
			fullVersion:  "6.0.22+LTS-T",
			searchString: "10.2",
			wantMatch:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesVersion(tt.fullVersion, tt.searchString)
			if got != tt.wantMatch {
				t.Errorf("matchesVersion(%s, %s) = %v, want %v",
					tt.fullVersion, tt.searchString, got, tt.wantMatch)
			}
		})
	}
}
