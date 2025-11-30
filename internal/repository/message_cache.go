package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-template-microservice/internal/models"
	"go-template-microservice/pkg/redis"

	"github.com/sirupsen/logrus"
)

type MessageCacheRepository interface {
	// CacheSentMessage caches the sent message with full details
	CacheSentMessage(ctx context.Context, message models.SentMessageCache) error
	// GetAllSentMessages retrieves all cached sent messages with limit
	GetAllSentMessages(ctx context.Context, limit int) ([]models.SentMessageCache, error)
}

type messageCacheRepository struct {
	redis  redis.IRedisInstance
	ttl    time.Duration
	logger *logrus.Logger
}

const (
	// sentMessageKeyPrefix is the prefix for sent message cache keys
	sentMessageKeyPrefix = "sent_message:"
)

func NewMessageCacheRepository(redis redis.IRedisInstance, ttl time.Duration, logger *logrus.Logger) MessageCacheRepository {
	return &messageCacheRepository{
		redis:  redis,
		ttl:    ttl,
		logger: logger,
	}
}

func (r *messageCacheRepository) CacheSentMessage(ctx context.Context, message models.SentMessageCache) error {
	key := fmt.Sprintf("%s%d", sentMessageKeyPrefix, message.MessageID)

	jsonData, err := json.Marshal(message)
	if err != nil {
		r.logger.WithError(err).WithField("messageID", message.MessageID).Error("Failed to marshal cache data")
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	err = r.redis.Client().Set(ctx, key, jsonData, r.ttl).Err()
	if err != nil {
		r.logger.WithError(err).WithField("messageID", message.MessageID).Error("Failed to cache sent message")
		return fmt.Errorf("failed to cache sent message: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"messageID":         message.MessageID,
		"externalMessageID": message.ExternalMessageID,
		"sentAt":            message.SentAt,
	}).Debug("Message cached successfully")

	return nil
}

// GetAllSentMessages retrieves all cached sent messages using SCAN command
// It returns up to 'limit' messages from the cache
func (r *messageCacheRepository) GetAllSentMessages(ctx context.Context, limit int) ([]models.SentMessageCache, error) {
	var messages []models.SentMessageCache
	var cursor uint64
	pattern := sentMessageKeyPrefix + "*"

	for {
		keys, nextCursor, err := r.redis.Client().Scan(ctx, cursor, pattern, int64(limit)).Result()
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan cache keys")
			return nil, fmt.Errorf("failed to scan cache keys: %w", err)
		}

		for _, key := range keys {
			if len(messages) >= limit {
				break
			}

			result, err := r.redis.Client().Get(ctx, key).Result()
			if err != nil {
				r.logger.WithError(err).WithField("key", key).Warn("Failed to get cached message, skipping")
				continue
			}

			var cacheData models.SentMessageCache
			if err := json.Unmarshal([]byte(result), &cacheData); err != nil {
				r.logger.WithError(err).WithField("key", key).Warn("Failed to unmarshal cache data, skipping")
				continue
			}

			messages = append(messages, cacheData)
		}

		cursor = nextCursor
		if cursor == 0 || len(messages) >= limit {
			break
		}
	}

	r.logger.WithField("count", len(messages)).Debug("Retrieved cached sent messages")
	return messages, nil
}
