package services

import (
	"context"
	"go-template-microservice/internal/models"
	"go-template-microservice/internal/repository"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type MessageScheduler interface {
	Start(c *fiber.Ctx)
	Stop(c *fiber.Ctx)
}

type messageScheduler struct {
	repo   repository.MessageRepository
	sender MessageSenderService
	cache  repository.MessageCacheRepository

	interval  time.Duration
	bacthSize int

	mu       sync.Mutex
	running  bool
	stopChan chan struct{}
	doneChan chan struct{}

	logger *logrus.Logger
}

func NewMessageScheduler(repo repository.MessageRepository, sender MessageSenderService, cache repository.MessageCacheRepository, interval time.Duration, batchSize int, logger *logrus.Logger) MessageScheduler {
	return &messageScheduler{
		repo:      repo,
		sender:    sender,
		cache:     cache,
		interval:  interval,
		bacthSize: batchSize,
		logger:    logger,
	}
}

func (s *messageScheduler) Start(c *fiber.Ctx) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.logger.Warn("Message scheduler is already running")
		return
	}

	s.running = true
	s.stopChan = make(chan struct{})
	s.doneChan = make(chan struct{})

	go s.loop(c.Context())
}

func (s *messageScheduler) Stop(c *fiber.Ctx) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	close(s.stopChan)
	s.mu.Unlock()

	<-s.doneChan
}

func (s *messageScheduler) loop(ctx context.Context) {
	defer func() {
		s.mu.Lock()
		s.running = false
		close(s.doneChan)
		s.mu.Unlock()
	}()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// run immediately on start
	s.tick(ctx)

	for {
		select {
		case <-ticker.C:
			s.tick(ctx)
		case <-s.stopChan:
			return
		}
	}
}

func (s *messageScheduler) tick(ctx context.Context) {
	messages, err := s.repo.GetUnsentMessages(s.bacthSize)
	if err != nil {
		s.logger.WithError(err).Error("Failed to retrieve unsent messages")
		return
	}

	for _, msg := range messages {
		resp, err := s.sender.Send(ctx, msg.To, msg.Content)
		if err != nil {
			s.logger.WithError(err).WithField("messageID", msg.ID).Error("Failed to send message")
			continue
		}
		sendAt := time.Now()
		err = s.repo.UpdateMessageStatus(msg.ID, models.StatusSent, &resp.MessageID, &sendAt)

		if s.cache != nil {
			cacheData := models.SentMessageCache{
				MessageID:         msg.ID,
				ExternalMessageID: resp.MessageID,
				To:                msg.To,
				Content:           msg.Content,
				SentAt:            sendAt,
			}
			cacheErr := s.cache.CacheSentMessage(ctx, cacheData)
			if cacheErr != nil {
				s.logger.WithError(cacheErr).WithField("messageID", msg.ID).Error("Failed to cache sent message")
			}
		}
	}
}
