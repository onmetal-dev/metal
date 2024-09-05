package templates

import (
	"errors"
	"testing"
)

func TestCls(t *testing.T) {
	tests := []struct {
		name      string
		baseClass string
		pairs     []interface{}
		expected  string
	}{
		{
			name:      "Empty base class, no pairs",
			baseClass: "",
			pairs:     []interface{}{},
			expected:  "",
		},
		{
			name:      "Base class only",
			baseClass: "base-class",
			pairs:     []interface{}{},
			expected:  "base-class",
		},
		{
			name:      "Single true boolean pair",
			baseClass: "base-class",
			pairs:     []interface{}{true, "true-class"},
			expected:  "base-class true-class",
		},
		{
			name:      "Single false boolean pair",
			baseClass: "base-class",
			pairs:     []interface{}{false, "false-class"},
			expected:  "base-class",
		},
		{
			name:      "Non-nil error pair",
			baseClass: "base-class",
			pairs:     []interface{}{error(errors.New("some error")), "error-class"},
			expected:  "base-class error-class",
		},
		{
			name:      "Nil error pair",
			baseClass: "base-class",
			pairs:     []interface{}{error(nil), "error-class"},
			expected:  "base-class",
		},
		{
			name:      "Multiple pairs",
			baseClass: "base-class",
			pairs:     []interface{}{true, "true-class", false, "false-class", "string", "string-class"},
			expected:  "base-class true-class string-class",
		},
		{
			name:      "Odd number of arguments",
			baseClass: "base-class",
			pairs:     []interface{}{true, "true-class", false},
			expected:  "base-class true-class",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cls(tt.baseClass, tt.pairs...)
			if result != tt.expected {
				t.Errorf("cls() = %v, want %v", result, tt.expected)
			}
		})
	}
}
