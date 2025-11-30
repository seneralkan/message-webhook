package middleware

import (
	"go-template-microservice/internal/constants"
	"go-template-microservice/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

func ValidationMiddleware(v validator.IValidation) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		ctx.Locals(constants.ValidatorContextKey, v)
		return ctx.Next()
	}
}
