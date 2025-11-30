package router

import (
	"fmt"
	"go-template-microservice/docs"
	"go-template-microservice/internal/handlers"
	"go-template-microservice/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type IRouter interface {
	RegisterRoutes(app *fiber.App)
}

type router struct {
	messageHandler handlers.MessageHandler
	logger         *logrus.Logger
}

// NewRouter
// @title go-template-microservice API
// @version 0.1
// @description The API provides go template-microservice service
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /
// swag init --parseDependency -g internal/router/router.go -o docs
func NewRouter(messageHandler handlers.MessageHandler, logger *logrus.Logger) IRouter {
	return &router{
		messageHandler: messageHandler,
		logger:         logger,
	}
}
func (r *router) RegisterRoutes(app *fiber.App) {
	app.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).JSON(utils.NewSuccessResponse(fiber.Map{"service": "up"}))
	})

	messageRouter := app.Group("/messages")
	r.RegisterMessageRoutes(messageRouter)
	r.Docs(app)
}

func (r *router) Docs(router fiber.Router) {
	router.Get("/documentation/*", docs.New(docs.Config{
		DeepLinking: true,
		URL:         fmt.Sprintf("/swagger/%s", docs.DefaultDocURL),
	}))
}
