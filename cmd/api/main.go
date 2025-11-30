package main

import (
	"go-template-microservice/internal/config"
	"go-template-microservice/internal/models"
	"go-template-microservice/pkg/redis"
	"go-template-microservice/pkg/sqlite"
	"go-template-microservice/pkg/validator"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	config := config.NewConfig()
	logger := logrus.New()

	logger.Info("Application Starting")

	// Initialize database with schema
	db, err := sqlite.NewSqliteInstanceWithSchemas(config.Database().Name, []string{
		models.GetMessageSchema(),
	})

	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	redis, err := redis.NewRedisInstance(config.Redis().Host, config.Redis().Port, config.Redis().Password, config.Redis().DB)

	if err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	router := CreateRouter(db, redis, config, logger)

	app := bootstrapApplication(&bootstrap{
		logger:    logger,
		validator: validator.BuildValidation(),
		configs:   config,
	})

	router.RegisterRoutes(app)

	go func() {
		if err := app.Listen(":" + config.Server().HttpPort); err != nil {
			logger.Errorf("Application Starting Error: %s", err.Error())
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	logger.Info("Info: Application Gracefully Shutting Down")
	if gShoutDown := app.Shutdown(); gShoutDown != nil {
		logger.Error(gShoutDown)
	}
}
