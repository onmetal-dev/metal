package form

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestForm struct {
	Name     string `form:"name" validate:"required"`
	Email    string `form:"email" validate:"required,email"`
	Username string `form:"username" validate:"required,lowercasealphanumhyphen"`
	EnvVars  string `form:"env_vars" validate:"dotenvformat"`
	Duration string `form:"duration" validate:"required,duration"`
}

func TestDecode(t *testing.T) {
	t.Run("Valid form data", func(t *testing.T) {
		formData := url.Values{
			"name":     {"John Doe"},
			"email":    {"john@example.com"},
			"username": {"john-doe"},
			"env_vars": {"KEY=value\nANOTHER_KEY=another_value"},
			"duration": {"1h"},
		}

		req, _ := http.NewRequest("POST", "/", nil)
		req.PostForm = formData

		var form TestForm
		fieldErrors, err := Decode(&form, req)
		if fieldErrors.NotNil() {
			for _, field := range fieldErrors.Fields() {
				assert.NoError(t, fieldErrors.Get(field), fmt.Sprintf("%s: %s", field, fieldErrors.Get(field).Error()))
			}
		}
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", form.Name)
		assert.Equal(t, "john@example.com", form.Email)
		assert.Equal(t, "john-doe", form.Username)
		assert.Equal(t, "KEY=value\nANOTHER_KEY=another_value", form.EnvVars)
		assert.Equal(t, "1h", form.Duration)
	})

	t.Run("Missing required field", func(t *testing.T) {
		formData := url.Values{
			"email":    {"john@example.com"},
			"username": {"john-doe"},
		}

		req, _ := http.NewRequest("POST", "/", nil)
		req.PostForm = formData

		var form TestForm
		fieldErrors, err := Decode(&form, req)

		require.Nil(t, err)
		require.True(t, fieldErrors.NotNil())
		require.Equal(t, "this is required", fieldErrors.Get("Name").Error())
	})

	t.Run("Invalid email", func(t *testing.T) {
		formData := url.Values{
			"name":     {"John Doe"},
			"email":    {"invalid-email"},
			"username": {"john-doe"},
		}

		req, _ := http.NewRequest("POST", "/", nil)
		req.PostForm = formData

		var form TestForm
		fieldErrors, err := Decode(&form, req)

		assert.Nil(t, err)
		assert.True(t, fieldErrors.NotNil())
		assert.NotNil(t, fieldErrors.Get("Email"))
	})
}

func TestInputValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"Zero value", 0, ""},
		{"Integer", 42, "42"},
		{"Negative integer", -10, "-10"},
		{"Float", 3.14, "3.14"},
		{"Float with trailing zeros", 2.5000, "2.5"},
		{"String", "hello", "hello"},
		{"Empty string", "", ""},
		{"Struct", struct{ Name string }{"John"}, "{John}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InputValue(tt.input)
			if result != tt.expected {
				t.Errorf("InputValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
