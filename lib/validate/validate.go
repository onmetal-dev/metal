// package validate provides a go-playground/validator instance to use with all of our custom validators
package validate

import (
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var validate *validator.Validate

// init initializes the global validator and adds any custom validators
func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("lowercasealphanumhyphen", isLowercaseAlphaNumHyphen)
	validate.RegisterValidation("dotenvformat", isDotenvFormat)
	validate.RegisterValidation("duration", isDuration)

}

var lowerCaseAlphaNumHyphenRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func isLowercaseAlphaNumHyphen(fl validator.FieldLevel) bool {
	return lowerCaseAlphaNumHyphenRegex.MatchString(fl.Field().String())
}

func isDotenvFormat(fl validator.FieldLevel) bool {
	input := fl.Field().String()
	_, err := godotenv.Parse(strings.NewReader(input))
	return err == nil
}

func isDuration(fl validator.FieldLevel) bool {
	_, err := time.ParseDuration(fl.Field().String())
	return err == nil
}

func Validator() *validator.Validate {
	return validate
}

func Struct(s interface{}) error {
	return validate.Struct(s)
}
