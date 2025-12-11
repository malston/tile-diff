// ABOUTME: Version resolution logic for fuzzy and exact matching.
// ABOUTME: Handles disambiguation and interactive selection when needed.
package pivnet

import (
	"fmt"
	"strings"
)

// ResolveResult holds the result of version resolution
type ResolveResult struct {
	Selected *Release  // The selected release (if resolved to one)
	Matches  []Release // All matching releases (if multiple found)
}

// Resolver handles version resolution
type Resolver struct {
	releases       []Release
	nonInteractive bool
}

// NewResolver creates a new version resolver
func NewResolver(releases []Release, nonInteractive bool) *Resolver {
	return &Resolver{
		releases:       releases,
		nonInteractive: nonInteractive,
	}
}

// Resolve resolves a version string to a specific release
func (r *Resolver) Resolve(version string) (*ResolveResult, error) {
	matches := []Release{}

	// Find all matching releases
	for _, rel := range r.releases {
		if matchesVersion(rel.Version, version) {
			matches = append(matches, rel)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no releases found matching version %s", version)
	}

	if len(matches) == 1 {
		// Single match - resolved
		return &ResolveResult{
			Selected: &matches[0],
			Matches:  matches,
		}, nil
	}

	// Multiple matches
	if r.nonInteractive {
		return nil, fmt.Errorf("multiple releases match version %s (use exact version string in non-interactive mode)", version)
	}

	// Interactive mode - return matches for user selection
	return &ResolveResult{
		Selected: nil,
		Matches:  matches,
	}, nil
}

// matchesVersion checks if a full version string matches a search string
func matchesVersion(fullVersion, searchString string) bool {
	// Exact match
	if fullVersion == searchString {
		return true
	}

	// Prefix match (fuzzy)
	return strings.HasPrefix(fullVersion, searchString)
}
