package main

import (
	"errors"
	"go-template-microservice/internal/config"
	"go-template-microservice/internal/constants"
	"go-template-microservice/internal/handlers"
	"go-template-microservice/internal/middleware"
	"go-template-microservice/internal/repository"
	"go-template-microservice/internal/router"
	"go-template-microservice/internal/services"
	"go-template-microservice/pkg/redis"
	"go-template-microservice/pkg/sqlite"
	"go-template-microservice/pkg/utils"
	"go-template-microservice/pkg/validator"
	"time"

	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type bootstrap struct {
	logger    *logrus.Logger
	validator validator.IValidation
	configs   config.IConfig
}

func bootstrapApplication(b *bootstrap) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(b.configs.Server().ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(b.configs.Server().WriteTimeout) * time.Second,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := utils.UnexpectedErrCode
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
				message = e.Message
			}
			errBag := utils.Error{Code: utils.UnexpectedErrCode, Message: message}
			return ctx.Status(code).JSON(utils.NewErrorResponse(ctx.Context(), errBag))
		},
	})
	b.registerMiddlewares(app)

	return app
}

func (b *bootstrap) registerMiddlewares(app *fiber.App) {
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Use(requestid.New(requestid.Config{
		ContextKey: constants.RequestIdKey,
	}))

	app.Use(middleware.ValidationMiddleware(b.validator))
}

func CreateRouter(
	db sqlite.ISqliteInstance,
	redis redis.IRedisInstance,
	cfg config.IConfig,
	l *logrus.Logger,
) router.IRouter {
	messageRepository := repository.NewMessageRepository(db, l)
	messageCacheRepository := repository.NewMessageCacheRepository(
		redis,
		time.Duration(cfg.Redis().TTLInSeconds)*time.Second,
		l,
	)
	messageSender := services.NewMessageSenderService(cfg.WebhookConfig().Url, cfg.WebhookConfig().AuthKey, l)
	messageScheduler := services.NewMessageScheduler(messageRepository, messageSender, messageCacheRepository, time.Duration(cfg.Scheduler().IntervalInSeconds)*time.Second, cfg.Scheduler().BatchSize, l)
	messageService := services.NewMessageService(messageRepository, messageCacheRepository, messageScheduler, l)

	messageHandler := handlers.NewMessageHandler(messageService, l)
	return router.NewRouter(messageHandler, l)
}
