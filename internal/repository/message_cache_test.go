package repository_test

import (
	"time"

	"go-template-microservice/internal/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessageCacheRepository", func() {
	BeforeEach(func() {
		// Clean up any existing cache entries before each test
		keys, err := mockRedis.Client().Keys(ctx, "sent_message:*").Result()
		Expect(err).NotTo(HaveOccurred())
		if len(keys) > 0 {
			err = mockRedis.Client().Del(ctx, keys...).Err()
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterEach(func() {
		// Clean up cache entries after each test
		keys, err := mockRedis.Client().Keys(ctx, "sent_message:*").Result()
		Expect(err).NotTo(HaveOccurred())
		if len(keys) > 0 {
			err = mockRedis.Client().Del(ctx, keys...).Err()
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("CacheSentMessage", func() {
		Context("when caching a valid message", func() {
			It("should cache the message successfully", func() {
				message := models.SentMessageCache{
					MessageID:         1,
					ExternalMessageID: "ext-123",
					To:                "+905551234567",
					Content:           "Test message content",
					SentAt:            time.Now(),
				}

				err := messageCacheRepository.CacheSentMessage(ctx, message)

				Expect(err).NotTo(HaveOccurred())

				// Verify the message was cached
				result, err := mockRedis.Client().Get(ctx, "sent_message:1").Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeEmpty())
				Expect(result).To(ContainSubstring("ext-123"))
			})
		})

		Context("when caching multiple messages", func() {
			It("should cache all messages successfully", func() {
				messages := []models.SentMessageCache{
					{
						MessageID:         1,
						ExternalMessageID: "ext-1",
						To:                "+905551111111",
						Content:           "Message 1",
						SentAt:            time.Now(),
					},
					{
						MessageID:         2,
						ExternalMessageID: "ext-2",
						To:                "+905552222222",
						Content:           "Message 2",
						SentAt:            time.Now(),
					},
					{
						MessageID:         3,
						ExternalMessageID: "ext-3",
						To:                "+905553333333",
						Content:           "Message 3",
						SentAt:            time.Now(),
					},
				}

				for _, msg := range messages {
					err := messageCacheRepository.CacheSentMessage(ctx, msg)
					Expect(err).NotTo(HaveOccurred())
				}

				// Verify all messages were cached
				keys, err := mockRedis.Client().Keys(ctx, "sent_message:*").Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(keys).To(HaveLen(3))
			})
		})

		Context("when updating an existing cached message", func() {
			It("should overwrite the previous cache entry", func() {
				message1 := models.SentMessageCache{
					MessageID:         1,
					ExternalMessageID: "ext-original",
					To:                "+905551234567",
					Content:           "Original content",
					SentAt:            time.Now(),
				}

				message2 := models.SentMessageCache{
					MessageID:         1,
					ExternalMessageID: "ext-updated",
					To:                "+905551234567",
					Content:           "Updated content",
					SentAt:            time.Now(),
				}

				err := messageCacheRepository.CacheSentMessage(ctx, message1)
				Expect(err).NotTo(HaveOccurred())

				err = messageCacheRepository.CacheSentMessage(ctx, message2)
				Expect(err).NotTo(HaveOccurred())

				// Verify only the updated message exists
				result, err := mockRedis.Client().Get(ctx, "sent_message:1").Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(ContainSubstring("ext-updated"))
				Expect(result).To(ContainSubstring("Updated content"))
			})
		})
	})

	Describe("GetAllSentMessages", func() {
		BeforeEach(func() {
			// Create some test cached messages
			messages := []models.SentMessageCache{
				{
					MessageID:         1,
					ExternalMessageID: "ext-1",
					To:                "+905551111111",
					Content:           "Message 1",
					SentAt:            time.Now().Add(-3 * time.Hour),
				},
				{
					MessageID:         2,
					ExternalMessageID: "ext-2",
					To:                "+905552222222",
					Content:           "Message 2",
					SentAt:            time.Now().Add(-2 * time.Hour),
				},
				{
					MessageID:         3,
					ExternalMessageID: "ext-3",
					To:                "+905553333333",
					Content:           "Message 3",
					SentAt:            time.Now().Add(-1 * time.Hour),
				},
			}

			for _, msg := range messages {
				err := messageCacheRepository.CacheSentMessage(ctx, msg)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		Context("when there are cached messages", func() {
			It("should return all cached messages up to the limit", func() {
				messages, err := messageCacheRepository.GetAllSentMessages(ctx, 10)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(HaveLen(3))
			})

			It("should respect the limit parameter", func() {
				messages, err := messageCacheRepository.GetAllSentMessages(ctx, 2)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(HaveLen(2))
			})

			It("should return messages with correct data", func() {
				messages, err := messageCacheRepository.GetAllSentMessages(ctx, 10)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(HaveLen(3))

				// Check that all messages have the required fields populated
				for _, msg := range messages {
					Expect(msg.MessageID).To(BeNumerically(">", 0))
					Expect(msg.ExternalMessageID).NotTo(BeEmpty())
					Expect(msg.To).NotTo(BeEmpty())
					Expect(msg.Content).NotTo(BeEmpty())
				}
			})
		})

		Context("when there are no cached messages", func() {
			BeforeEach(func() {
				// Clean up all cache entries
				keys, err := mockRedis.Client().Keys(ctx, "sent_message:*").Result()
				Expect(err).NotTo(HaveOccurred())
				if len(keys) > 0 {
					err = mockRedis.Client().Del(ctx, keys...).Err()
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("should return an empty slice", func() {
				messages, err := messageCacheRepository.GetAllSentMessages(ctx, 10)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(BeEmpty())
			})
		})

		Context("when limit is 1", func() {
			It("should return only one message", func() {
				messages, err := messageCacheRepository.GetAllSentMessages(ctx, 1)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(HaveLen(1))
			})
		})
	})

	Describe("Cache TTL", func() {
		Context("when checking cache expiration", func() {
			It("should set TTL on cached messages", func() {
				message := models.SentMessageCache{
					MessageID:         999,
					ExternalMessageID: "ext-ttl-test",
					To:                "+905551234567",
					Content:           "TTL test message",
					SentAt:            time.Now(),
				}

				err := messageCacheRepository.CacheSentMessage(ctx, message)
				Expect(err).NotTo(HaveOccurred())

				// Check that TTL is set
				ttl, err := mockRedis.Client().TTL(ctx, "sent_message:999").Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(ttl).To(BeNumerically(">", 0))
				Expect(ttl).To(BeNumerically("<=", cacheTTL))
			})
		})
	})
})
