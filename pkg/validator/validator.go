package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// Validate runs struct validation and, on failure, returns an error whose
// Error() is a human-readable summary (e.g. "Name is required; Price must
// be at least 0") instead of go-playground's Go-struct-shaped message —
// every handler forwards this string straight to API clients as-is.
func (v *Validator) Validate(i interface{}) error {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	var fieldErrs validator.ValidationErrors
	if errors.As(err, &fieldErrs) {
		return errors.New(friendlyMessage(fieldErrs))
	}
	return err
}

func friendlyMessage(fieldErrs validator.ValidationErrors) string {
	messages := make([]string, 0, len(fieldErrs))
	for _, fe := range fieldErrs {
		messages = append(messages, fieldMessage(fe))
	}
	return strings.Join(messages, "; ")
}

func fieldMessage(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid ID", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
