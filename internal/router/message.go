package router

import (
	_ "go-template-microservice/internal/resources/response"
	_ "go-template-microservice/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

func (r *router) RegisterMessageRoutes(router fiber.Router) {
	r.RegisterMessageStartSchedulerRoute(router)
	r.RegisterMessageStopSchedulerRoute(router)
	r.RegisterMessageListSentMessagesRoute(router)
}

// RegisterMessageListSentMessagesRoute registers the route to list sent messages
// @Summary List Sent Messages
// @Description Retrieves a list of sent messages
// @Tags Messages
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of messages to retrieve" default(10)
// @Success 200 {object} response.SentMessagesResponse
// @Failure 500 {object} utils.HTTPErrorResponse
// @Router /messages/sent [get]
func (r *router) RegisterMessageListSentMessagesRoute(router fiber.Router) {
	router.Get("/sent", r.messageHandler.ListSentMessages)
}

// RegisterMessageStartSchedulerRoute registers the route to start the message scheduler
// @Summary Start Message Scheduler
// @Description Starts the message sending scheduler
// @Tags Messages
// @Accept json
// @Produce json
// @Success 200 {object} utils.HTTPSuccessResponse
// @Failure 500 {object} utils.HTTPErrorResponse
// @Router /messages/start [post]
func (r *router) RegisterMessageStartSchedulerRoute(router fiber.Router) {
	router.Post("/start", r.messageHandler.StartScheduler)
}

// RegisterMessageStopSchedulerRoute registers the route to stop the message scheduler
// @Summary Stop Message Scheduler
// @Description Stops the message sending scheduler
// @Tags Messages
// @Accept json
// @Produce json
// @Success 200 {object} utils.HTTPSuccessResponse
// @Failure 500 {object} utils.HTTPErrorResponse
// @Router /messages/stop [post]
func (r *router) RegisterMessageStopSchedulerRoute(router fiber.Router) {
	router.Post("/stop", r.messageHandler.StopScheduler)
}
