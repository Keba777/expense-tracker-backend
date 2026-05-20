package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	v *validator.Validate
}

func New() *Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld interface{ Tag(string) string }) string {
		name := fld.Tag("json")
		if name == "-" {
			return ""
		}
		if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}
		return name
	})

	return &Validator{v: v}
}

func (val *Validator) Validate(s interface{}) error {
	if err := val.v.Struct(s); err != nil {
		var errs validator.ValidationErrors
		if ok := strings.Contains(err.Error(), "Error"); ok {
			if ve, ok2 := err.(validator.ValidationErrors); ok2 {
				errs = ve
			}
		}
		if errs != nil {
			messages := make([]string, 0, len(errs))
			for _, e := range errs {
				messages = append(messages, fieldError(e))
			}
			return fmt.Errorf("%s", strings.Join(messages, "; "))
		}
		return err
	}
	return nil
}

func fieldError(e validator.FieldError) string {
	field := e.Field()
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	default:
		return fmt.Sprintf("%s failed validation: %s", field, e.Tag())
	}
}
