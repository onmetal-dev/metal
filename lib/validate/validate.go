// package validate provides a go-playground/validator instance to use with all of our custom validators
package validate

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// init initializes the global validator and adds any custom validators
func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("lowercasealphanumhyphen", isLowercaseAlphaNumHyphen)
}

var lowerCaseAlphaNumHyphenRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func isLowercaseAlphaNumHyphen(fl validator.FieldLevel) bool {
	return lowerCaseAlphaNumHyphenRegex.MatchString(fl.Field().String())
}

func Validator() *validator.Validate {
	return validate
}

func Struct(s interface{}) error {
	return validate.Struct(s)
}
