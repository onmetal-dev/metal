package templates

import (
	"errors"
	"testing"
)

func TestCls(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected string
	}{
		{
			name:     "Empty base class, no pairs",
			args:     []interface{}{""},
			expected: "",
		},
		{
			name:     "Base class only",
			args:     []interface{}{"base-class"},
			expected: "base-class",
		},
		{
			name:     "Single true boolean pair",
			args:     []interface{}{"base-class", true, "true-class"},
			expected: "base-class true-class",
		},
		{
			name:     "Single false boolean pair",
			args:     []interface{}{"base-class", false, "false-class"},
			expected: "base-class",
		},
		{
			name:     "Non-nil error pair",
			args:     []interface{}{"base-class", error(errors.New("some error")), "error-class"},
			expected: "base-class error-class",
		},
		{
			name:     "Nil error pair",
			args:     []interface{}{"base-class", error(nil), "error-class"},
			expected: "base-class",
		},
		{
			name:     "Multiple pairs",
			args:     []interface{}{"base-class", true, "true-class", false, "false-class", "string", "string-class"},
			expected: "base-class true-class string string-class",
		},
		{
			name:     "Odd number of arguments",
			args:     []interface{}{"base-class", true, "true-class", false},
			expected: "base-class true-class",
		},
		{
			name:     "String-only arguments",
			args:     []interface{}{"base-class", "additional-class", "another-class"},
			expected: "base-class additional-class another-class",
		},
		{
			name:     "Mixed string and pair arguments",
			args:     []interface{}{"base-class", "additional-class", true, "true-class", "another-class", false, "false-class"},
			expected: "base-class additional-class true-class another-class",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cls(tt.args...)
			if result != tt.expected {
				t.Errorf("cls() = %v, want %v", result, tt.expected)
			}
		})
	}
}
