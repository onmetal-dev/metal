package validate

import (
	"testing"
)

func TestValidator(t *testing.T) {
	v := Validator()
	if v == nil {
		t.Error("Validator() returned nil")
	}
}

func TestLowercaseAlphaNumHyphen(t *testing.T) {
	type TestStruct struct {
		Field string `validate:"lowercasealphanumhyphen"`
	}

	v := Validator()

	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"valid", "abc-123", true},
		{"uppercase", "ABC-123", false},
		{"with underscore", "abc_123", false},
		{"with space", "abc 123", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := TestStruct{Field: tc.input}
			err := v.Struct(ts)

			if tc.valid && err != nil {
				t.Errorf("Expected valid input, but got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Errorf("Expected invalid input, but got no error")
			}
		})
	}
}

func TestDotenvFormat(t *testing.T) {
	type TestStruct struct {
		Field string `validate:"dotenvformat"`
	}

	v := Validator()

	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"valid single line", "KEY=value", true},
		{"valid multiple lines", "KEY1=value1\nKEY2=value2", true},
		{"valid with empty value", "KEY=", true},
		{"invalid format", "INVALID LINE\n", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := TestStruct{Field: tc.input}
			err := v.Struct(ts)

			if tc.valid && err != nil {
				t.Errorf("Test case '%s': Expected valid input '%s', but got error: %v", tc.name, tc.input, err)
			}
			if !tc.valid && err == nil {
				t.Errorf("Test case '%s': Expected invalid input '%s', but got no error", tc.name, tc.input)
			}
		})
	}
}
