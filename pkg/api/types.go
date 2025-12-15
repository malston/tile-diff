// ABOUTME: Defines data structures for Ops Manager API responses.
// ABOUTME: Maps JSON property format to Go structs for comparison.
package api

// Property represents a single property from Ops Manager API
type Property struct {
	Type           string      `json:"type"`
	Configurable   bool        `json:"configurable"`
	Credential     bool        `json:"credential"`
	Value          interface{} `json:"value"`
	Optional       bool        `json:"optional"`
	SelectedOption *string     `json:"selected_option,omitempty"`
}

// PropertiesResponse represents the API response for product properties
type PropertiesResponse struct {
	Properties map[string]Property `json:"properties"`
}

// StagedProduct represents a staged product in Ops Manager
type StagedProduct struct {
	GUID string `json:"guid"`
	Type string `json:"type"`
}
