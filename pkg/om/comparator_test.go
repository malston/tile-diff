// ABOUTME: Tests for configuration comparison logic.
// ABOUTME: Verifies detection of added, removed, and changed properties.
package om

import (
	"testing"
)

func TestCompareConfigs(t *testing.T) {
	tests := []struct {
		name string
		old  *ProductConfig
		new  *ProductConfig
		want *ConfigComparison
	}{
		{
			name: "no changes",
			old: &ProductConfig{
				ProductName: "test-product",
				ProductProperties: map[string]PropertyValue{
					".properties.foo": {Value: "bar"},
					".properties.baz": {Value: 123},
				},
			},
			new: &ProductConfig{
				ProductName: "test-product",
				ProductProperties: map[string]PropertyValue{
					".properties.foo": {Value: "bar"},
					".properties.baz": {Value: 123},
				},
			},
			want: &ConfigComparison{
				Added:   []ConfigChange{},
				Removed: []ConfigChange{},
				Changed: []ConfigChange{},
			},
		},
		{
			name: "property added",
			old: &ProductConfig{
				ProductName: "test-product",
				ProductProperties: map[string]PropertyValue{
					".properties.foo": {Value: "bar"},
				},
			},
			new: &ProductConfig{
				ProductName: "test-product",
				ProductProperties: map[string]PropertyValue{
					".properties.foo": {Value: "bar"},
					".properties.new": {Value: "value"},
				},
			},
			want: &ConfigComparison{
				Added: []ConfigChange{
					{
						PropertyName: ".properties.new",
						ChangeType:   "added",
						Description:  "New configurable property",
						NewValue:     "value",
					},
				},
				Removed: []ConfigChange{},
				Changed: []ConfigChange{},
			},
		},
		{
			name: "property removed",
			old: &ProductConfig{
				ProductName: "test-product",
				ProductProperties: map[string]PropertyValue{
					".properties.foo": {Value: "bar"},
					".properties.old": {Value: "value"},
				},
			},
			new: &ProductConfig{
				ProductName: "test-product",
				ProductProperties: map[string]PropertyValue{
					".properties.foo": {Value: "bar"},
				},
			},
			want: &ConfigComparison{
				Added: []ConfigChange{},
				Removed: []ConfigChange{
					{
						PropertyName: ".properties.old",
						ChangeType:   "removed",
						Description:  "Property removed",
						OldValue:     "value",
					},
				},
				Changed: []ConfigChange{},
			},
		},
		{
			name: "property value changed",
			old: &ProductConfig{
				ProductName: "test-product",
				ProductProperties: map[string]PropertyValue{
					".properties.foo": {Value: "old-value"},
				},
			},
			new: &ProductConfig{
				ProductName: "test-product",
				ProductProperties: map[string]PropertyValue{
					".properties.foo": {Value: "new-value"},
				},
			},
			want: &ConfigComparison{
				Added: []ConfigChange{},
				Removed: []ConfigChange{},
				Changed: []ConfigChange{
					{
						PropertyName: ".properties.foo",
						ChangeType:   "changed",
						Description:  "Default value changed",
						OldValue:     "old-value",
						NewValue:     "new-value",
					},
				},
			},
		},
		{
			name: "multiple changes",
			old: &ProductConfig{
				ProductName: "test-product",
				ProductProperties: map[string]PropertyValue{
					".properties.unchanged": {Value: "same"},
					".properties.changed":   {Value: "old"},
					".properties.removed":   {Value: "gone"},
				},
			},
			new: &ProductConfig{
				ProductName: "test-product",
				ProductProperties: map[string]PropertyValue{
					".properties.unchanged": {Value: "same"},
					".properties.changed":   {Value: "new"},
					".properties.added":     {Value: "here"},
				},
			},
			want: &ConfigComparison{
				Added: []ConfigChange{
					{
						PropertyName: ".properties.added",
						ChangeType:   "added",
						Description:  "New configurable property",
						NewValue:     "here",
					},
				},
				Removed: []ConfigChange{
					{
						PropertyName: ".properties.removed",
						ChangeType:   "removed",
						Description:  "Property removed",
						OldValue:     "gone",
					},
				},
				Changed: []ConfigChange{
					{
						PropertyName: ".properties.changed",
						ChangeType:   "changed",
						Description:  "Default value changed",
						OldValue:     "old",
						NewValue:     "new",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareConfigs(tt.old, tt.new)

			// Check counts
			if len(got.Added) != len(tt.want.Added) {
				t.Errorf("Added count = %d, want %d", len(got.Added), len(tt.want.Added))
			}
			if len(got.Removed) != len(tt.want.Removed) {
				t.Errorf("Removed count = %d, want %d", len(got.Removed), len(tt.want.Removed))
			}
			if len(got.Changed) != len(tt.want.Changed) {
				t.Errorf("Changed count = %d, want %d", len(got.Changed), len(tt.want.Changed))
			}

			// Verify added properties
			for i, want := range tt.want.Added {
				if i >= len(got.Added) {
					break
				}
				if got.Added[i].PropertyName != want.PropertyName {
					t.Errorf("Added[%d].PropertyName = %s, want %s", i, got.Added[i].PropertyName, want.PropertyName)
				}
			}

			// Verify removed properties
			for i, want := range tt.want.Removed {
				if i >= len(got.Removed) {
					break
				}
				if got.Removed[i].PropertyName != want.PropertyName {
					t.Errorf("Removed[%d].PropertyName = %s, want %s", i, got.Removed[i].PropertyName, want.PropertyName)
				}
			}

			// Verify changed properties
			for i, want := range tt.want.Changed {
				if i >= len(got.Changed) {
					break
				}
				if got.Changed[i].PropertyName != want.PropertyName {
					t.Errorf("Changed[%d].PropertyName = %s, want %s", i, got.Changed[i].PropertyName, want.PropertyName)
				}
			}
		})
	}
}
