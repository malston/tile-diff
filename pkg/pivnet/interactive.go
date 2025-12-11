// ABOUTME: Interactive prompts for user selection.
// ABOUTME: Provides terminal-based selection UI for releases and product files.
package pivnet

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

// PromptForRelease prompts user to select a release
func PromptForRelease(releases []Release) (*Release, error) {
	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases available")
	}

	if len(releases) == 1 {
		return &releases[0], nil
	}

	options := buildReleaseOptions(releases)

	var selected string
	prompt := &survey.Select{
		Message: "Multiple releases found. Select version:",
		Options: options,
		Default: options[0],
	}

	err := survey.AskOne(prompt, &selected)
	if err != nil {
		return nil, err
	}

	// Find selected release
	for i, opt := range options {
		if opt == selected {
			return &releases[i], nil
		}
	}

	return nil, fmt.Errorf("selection failed")
}

// PromptForProductFile prompts user to select a product file
func PromptForProductFile(files []ProductFile) (*ProductFile, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no product files available")
	}

	if len(files) == 1 {
		return &files[0], nil
	}

	options := buildProductFileOptions(files)

	// Try to find "TAS for VMs" as default
	defaultIdx := 0
	for i, f := range files {
		if f.Name == "TAS for VMs" {
			defaultIdx = i
			break
		}
	}

	var selected string
	prompt := &survey.Select{
		Message: "Multiple product files found. Select file:",
		Options: options,
		Default: options[defaultIdx],
	}

	err := survey.AskOne(prompt, &selected)
	if err != nil {
		return nil, err
	}

	// Find selected file
	for i, opt := range options {
		if opt == selected {
			return &files[i], nil
		}
	}

	return nil, fmt.Errorf("selection failed")
}

// PromptForEULA prompts user to accept EULA
func PromptForEULA(productSlug, version, eulaURL string) (bool, error) {
	fmt.Printf("\nYou must accept the EULA to download %s %s\n", productSlug, version)
	fmt.Printf("View EULA: %s\n\n", eulaURL)

	var accepted bool
	prompt := &survey.Confirm{
		Message: "Accept EULA? (will be remembered for this product)",
		Default: false,
	}

	err := survey.AskOne(prompt, &accepted)
	if err != nil {
		return false, err
	}

	return accepted, nil
}

// buildReleaseOptions creates display options for releases
func buildReleaseOptions(releases []Release) []string {
	options := make([]string, len(releases))
	for i, r := range releases {
		if i == 0 {
			options[i] = r.Version + " (latest)"
		} else {
			options[i] = r.Version
		}
	}
	return options
}

// buildProductFileOptions creates display options for product files
func buildProductFileOptions(files []ProductFile) []string {
	options := make([]string, len(files))
	for i, f := range files {
		sizeStr := formatBytes(f.Size)
		recommended := ""
		if f.Name == "TAS for VMs" {
			recommended = " [Recommended]"
		}
		options[i] = fmt.Sprintf("%s (%s)%s", f.Name, sizeStr, recommended)
	}
	return options
}
