// ABOUTME: Defines data structures for om config-template output.
// ABOUTME: Maps YAML configuration templates to Go structs.
package om

// PropertyValue represents a property value in product.yml
type PropertyValue struct {
	Value interface{} `yaml:"value,omitempty"`
}

// ProductConfig represents the parsed product.yml template from om config-template
type ProductConfig struct {
	ProductName       string                   `yaml:"product-name"`
	ProductProperties map[string]PropertyValue `yaml:"product-properties"`
}
