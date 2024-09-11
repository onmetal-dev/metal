// package forms helps parse and validate form data
package form

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/onmetal-dev/metal/lib/validate"
	"github.com/samber/lo"
)

type FieldErrors struct {
	errors map[string]error
}

func (e FieldErrors) NotNil() bool {
	return e.errors != nil
}

func (e *FieldErrors) Set(field string, err error) {
	if e.errors == nil {
		e.errors = make(map[string]error)
	}
	e.errors[field] = err
}

func (e FieldErrors) Get(field string) error {
	if e.errors == nil {
		return nil
	}
	return e.errors[field]
}

func (e FieldErrors) Fields() []string {
	if e.errors == nil {
		return []string{}
	}
	return lo.MapToSlice(e.errors, func(key string, _ error) string {
		return key
	})
}

// DecodeFormData parses form data from a request. It uses the go-playground/form
// package to decode the form data into the struct, and then the
// go-playground/validator package to validate the struct.
// It returns two error types: FieldErrors for specific field-level errors, and a
// generic error for any other errors that occurred during parsing or validation.
func Decode[T any](formData *T, r *http.Request) (FieldErrors, error) {
	var fieldErrors FieldErrors
	if err := r.ParseForm(); err != nil {
		return fieldErrors, err
	}
	decoder := form.NewDecoder()
	if err := decoder.Decode(formData, r.Form); err != nil {
		return fieldErrors, err
	}

	if err := validate.Struct(formData); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			switch err.Tag() {
			case "required":
				fieldErrors.Set(field, errors.New("this is required"))
			case "lowercasealphanumhyphen":
				fieldErrors.Set(field, errors.New("must consist of lowercase alphanumeric characters and/or hyphens"))
			case "dotenvformat":
				fieldErrors.Set(field, errors.New("must be in dotenv format"))
			default:
				fieldErrors.Set(field, err)
			}
		}
	}
	return fieldErrors, nil
}

// InputValue converts a value to a string that can be used as the value of an input element.
func InputValue[T any](v T) string {
	if reflect.ValueOf(v).IsZero() {
		return ""
	}

	switch val := any(v).(type) {
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		s := fmt.Sprintf("%f", val)
		if strings.Contains(s, ".") {
			s = strings.TrimRight(s, "0")
			s = strings.TrimRight(s, ".")
		}
		return s
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}
