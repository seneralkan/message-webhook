package utils

import (
	"context"
	"go-template-microservice/internal/constants"
	"go-template-microservice/pkg/validator"
)

func Validator(ctx context.Context, i interface{}) map[string]string {
	v := ctx.Value(constants.ValidatorContextKey).(validator.IValidation)
	if errs := v.Validate(i); len(errs) > 0 {
		return errs
	}
	return nil
}
