package handlers

import (
	"go-template-microservice/internal/resources/request"
	"go-template-microservice/internal/services"
	"go-template-microservice/pkg/utils"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type MessageHandler interface {
	StartScheduler(c *fiber.Ctx) error
	StopScheduler(c *fiber.Ctx) error
	ListSentMessages(c *fiber.Ctx) error
}

type messageHandler struct {
	messageService services.MessageService
	logger         *logrus.Logger
}

func NewMessageHandler(mmessageService services.MessageService, logger *logrus.Logger) MessageHandler {
	return &messageHandler{
		messageService: mmessageService,
	}
}

func (h *messageHandler) StartScheduler(c *fiber.Ctx) error {
	h.messageService.StartScheduler(c)
	return c.Status(http.StatusOK).JSON(utils.NewSuccessResponse(fiber.Map{"state": "started"}))
}

func (h *messageHandler) StopScheduler(c *fiber.Ctx) error {
	h.messageService.StopScheduler(c)
	return c.Status(http.StatusOK).JSON(utils.NewSuccessResponse(fiber.Map{"state": "stopped"}))
}

func (h *messageHandler) ListSentMessages(c *fiber.Ctx) error {
	var req request.ListSentMessagesRequest
	if err := c.QueryParser(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse ListSentMessagesRequest")
		return c.Status(fiber.StatusUnprocessableEntity).JSON((utils.NewBodyParserErrorResponse()))
	}

	if err := utils.Validator(c.Context(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.NewValidationErrorResponse(err))
	}

	const defaultLimit = 10
	limit := c.QueryInt("limit", defaultLimit)
	if limit <= 0 {
		limit = defaultLimit
	}

	messages, err := h.messageService.ListSentMessages(c, limit)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(utils.NewSuccessResponse(messages))
}
