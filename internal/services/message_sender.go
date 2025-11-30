package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-template-microservice/internal/resources/response"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type MessageSenderService interface {
	Send(ctx context.Context, to, content string) (*response.WebhookResponse, error)
}

type messageSenderService struct {
	client     *http.Client
	webHookURL string
	authKey    string
	logger     *logrus.Logger
}

func NewMessageSenderService(webHookURL, authKey string, logger *logrus.Logger) MessageSenderService {
	return &messageSenderService{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		webHookURL: webHookURL,
		authKey:    authKey,
		logger:     logger,
	}
}

func (s *messageSenderService) Send(ctx context.Context, to, content string) (*response.WebhookResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"to":      to,
		"content": content,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", s.webHookURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ins-auth-key", s.authKey)

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.WithError(err).Error("Failed to send message")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		logrus.WithField("status_code", resp.StatusCode).Error("Failed to send message, non-202 response")
		return nil, fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	var wResp response.WebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&wResp); err != nil {
		s.logger.WithError(err).Error("Failed to decode webhook response")
		return nil, err
	}

	return &wResp, nil
}
