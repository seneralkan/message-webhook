package services

import (
	"sort"
	"time"

	"go-template-microservice/internal/repository"
	"go-template-microservice/internal/resources/response"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type MessageService interface {
	StartScheduler(c *fiber.Ctx)
	StopScheduler(c *fiber.Ctx)
	ListSentMessages(ctx *fiber.Ctx, limit int) ([]response.SentMessageResponse, error)
}

type sortableMessage struct {
	response response.SentMessageResponse
	sentAt   time.Time
}

type messageService struct {
	repo      repository.MessageRepository
	cacheRepo repository.MessageCacheRepository
	scheduler MessageScheduler
	logger    *logrus.Logger
}

func NewMessageService(
	repo repository.MessageRepository,
	cacheRepo repository.MessageCacheRepository,
	scheduler MessageScheduler,
	logger *logrus.Logger,
) MessageService {
	return &messageService{
		repo:      repo,
		cacheRepo: cacheRepo,
		scheduler: scheduler,
		logger:    logger,
	}
}

func (s *messageService) StartScheduler(c *fiber.Ctx) {
	s.scheduler.Start(c)
}

func (s *messageService) StopScheduler(c *fiber.Ctx) {
	s.scheduler.Stop(c)
}

// ListSentMessages returns sent messages sorted by sentAt descending (newest first).
// It combines results from cache and database, ensuring consistent ordering.
func (s *messageService) ListSentMessages(ctx *fiber.Ctx, limit int) ([]response.SentMessageResponse, error) {
	var sortable []sortableMessage

	// First try to get from cache
	cachedMessages, err := s.cacheRepo.GetAllSentMessages(ctx.Context(), limit)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get messages from cache, falling back to database")
	} else if len(cachedMessages) > 0 {
		s.logger.WithField("count", len(cachedMessages)).Debug("Retrieved messages from cache")

		for _, cached := range cachedMessages {
			sortable = append(sortable, sortableMessage{
				response: response.SentMessageResponse{
					MessageID:         cached.MessageID,
					ExternalMessageID: cached.ExternalMessageID,
					To:                cached.To,
					Content:           cached.Content,
					SentAt:            cached.SentAt.Format("2006-01-02 15:04:05"),
				},
				sentAt: cached.SentAt,
			})
		}
	}

	if len(sortable) >= limit {
		// Sort by sentAt descending before returning
		sort.Slice(sortable, func(i, j int) bool {
			return sortable[i].sentAt.After(sortable[j].sentAt)
		})
		responses := make([]response.SentMessageResponse, limit)
		for i := 0; i < limit; i++ {
			responses[i] = sortable[i].response
		}
		return responses, nil
	}

	remainingLimit := limit - len(sortable)
	s.logger.WithField("remainingLimit", remainingLimit).Debug("Fetching additional messages from database")

	dbMessages, err := s.repo.GetSentMessages(remainingLimit)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get sent messages from database")
		// If we have some cached responses, return them instead of failing
		if len(sortable) > 0 {
			// Sort by sentAt descending before returning
			sort.Slice(sortable, func(i, j int) bool {
				return sortable[i].sentAt.After(sortable[j].sentAt)
			})
			responses := make([]response.SentMessageResponse, len(sortable))
			for i, s := range sortable {
				responses[i] = s.response
			}
			return responses, nil
		}
		return nil, err
	}

	// Build a set of already cached message IDs to avoid duplicates
	cachedIDs := make(map[int64]bool)
	for _, s := range sortable {
		cachedIDs[s.response.MessageID] = true
	}

	// Add messages from DB that aren't already in cache
	for _, msg := range dbMessages {
		if cachedIDs[msg.ID] {
			continue // Skip duplicates
		}

		sortable = append(sortable, sortableMessage{
			response: response.SentMessageResponse{
				MessageID:         msg.ID,
				ExternalMessageID: msg.ExternalMessageID,
				To:                msg.To,
				Content:           msg.Content,
				SentAt:            msg.SentAt.Format("2006-01-02 15:04:05"),
			},
			sentAt: msg.SentAt,
		})

		if len(sortable) >= limit {
			break
		}
	}

	// Sort combined results by sentAt descending (newest first)
	sort.Slice(sortable, func(i, j int) bool {
		return sortable[i].sentAt.After(sortable[j].sentAt)
	})

	responses := make([]response.SentMessageResponse, len(sortable))
	for i, s := range sortable {
		responses[i] = s.response
	}

	return responses, nil
}
