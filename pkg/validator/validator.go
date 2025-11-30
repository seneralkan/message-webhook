package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type IValidation interface {
	Validate(i interface{}) map[string]string
	RegisterValidation(fieldName string, registerFunc func(fl validator.FieldLevel) bool) error
}

type validate struct {
	v *validator.Validate
}

func BuildValidation() IValidation {
	v := validator.New()
	return &validate{v: v}
}

func (v *validate) Validate(i interface{}) map[string]string {
	validationMessages := make(map[string]string)
	if err := v.v.Struct(i); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validationMessages[err.Field()] = fmt.Sprintf("field validation for %s failed on the %s tag", err.Field(), err.Tag())
		}
	}
	return validationMessages
}

func (v *validate) RegisterValidation(fieldName string, registerFunc func(fl validator.FieldLevel) bool) error {
	return v.v.RegisterValidation(fieldName, registerFunc)
}
