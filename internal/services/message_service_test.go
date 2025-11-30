package services_test

import (
	"errors"
	"time"

	"go-template-microservice/internal/models"
	"go-template-microservice/internal/services"

	"github.com/gofiber/fiber/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"
)

var _ = Describe("MessageService", func() {
	BeforeEach(func() {
		_, err := sqliteInst.Database().Exec("DELETE FROM messages")
		Expect(err).NotTo(HaveOccurred())

		keys, err := redisInst.Client().Keys(ctx, "sent_message:*").Result()
		Expect(err).NotTo(HaveOccurred())
		if len(keys) > 0 {
			err = redisInst.Client().Del(ctx, keys...).Err()
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("ListSentMessages", func() {
		Context("when there are messages in cache and database", func() {
			It("should set up repositories correctly", func() {
				msg1, err := messageRepository.CreateMessage("+905551111111", "DB Message 1")
				Expect(err).NotTo(HaveOccurred())
				extID1 := "db-ext-1"
				sentAt1 := time.Now().Add(-2 * time.Hour)
				err = messageRepository.UpdateMessageStatus(msg1.ID, models.StatusSent, &extID1, &sentAt1)
				Expect(err).NotTo(HaveOccurred())

				msg2, err := messageRepository.CreateMessage("+905552222222", "DB Message 2")
				Expect(err).NotTo(HaveOccurred())
				extID2 := "db-ext-2"
				sentAt2 := time.Now().Add(-1 * time.Hour)
				err = messageRepository.UpdateMessageStatus(msg2.ID, models.StatusSent, &extID2, &sentAt2)
				Expect(err).NotTo(HaveOccurred())

				cacheData := models.SentMessageCache{
					MessageID:         msg1.ID,
					ExternalMessageID: extID1,
					To:                msg1.To,
					Content:           msg1.Content,
					SentAt:            sentAt1,
				}
				err = messageCacheRepository.CacheSentMessage(ctx, cacheData)
				Expect(err).NotTo(HaveOccurred())

				service := services.NewMessageService(
					messageRepository,
					messageCacheRepository,
					nil,
					logger,
				)

				_ = service
			})
		})

		Context("with mocked repositories", func() {
			It("should return cached messages when available", func() {
				now := time.Now()
				cachedMessages := []models.SentMessageCache{
					{
						MessageID:         1,
						ExternalMessageID: "cache-ext-1",
						To:                "+905551111111",
						Content:           "Cached Message 1",
						SentAt:            now.Add(-1 * time.Hour),
					},
					{
						MessageID:         2,
						ExternalMessageID: "cache-ext-2",
						To:                "+905552222222",
						Content:           "Cached Message 2",
						SentAt:            now.Add(-2 * time.Hour),
					},
				}

				messageCacheMock.EXPECT().
					GetAllSentMessages(gomock.Any(), 10).
					Return(cachedMessages, nil).
					Times(1)

				// Since cache has 2 messages and limit is 10, service will try to get 8 more from DB
				messageRepoMock.EXPECT().
					GetSentMessages(8).
					Return([]models.Message{}, nil).
					Times(1)

				service := services.NewMessageService(
					messageRepoMock,
					messageCacheMock,
					nil,
					logger,
				)

				// Create a fiber context for testing
				app := fiber.New()
				fiberCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				defer app.ReleaseCtx(fiberCtx)

				responses, err := service.ListSentMessages(fiberCtx, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(responses).To(HaveLen(2))
				// Should be sorted by sentAt descending (newest first)
				Expect(responses[0].MessageID).To(Equal(int64(1))) // -1 hour (newer)
				Expect(responses[1].MessageID).To(Equal(int64(2))) // -2 hours (older)
			})

			It("should sort cached messages by sentAt descending", func() {
				now := time.Now()
				// Return messages in arbitrary order (simulating Redis SCAN)
				cachedMessages := []models.SentMessageCache{
					{
						MessageID:         3,
						ExternalMessageID: "cache-ext-3",
						To:                "+905553333333",
						Content:           "Oldest Message",
						SentAt:            now.Add(-3 * time.Hour), // oldest
					},
					{
						MessageID:         1,
						ExternalMessageID: "cache-ext-1",
						To:                "+905551111111",
						Content:           "Newest Message",
						SentAt:            now.Add(-1 * time.Hour), // newest
					},
					{
						MessageID:         2,
						ExternalMessageID: "cache-ext-2",
						To:                "+905552222222",
						Content:           "Middle Message",
						SentAt:            now.Add(-2 * time.Hour), // middle
					},
				}

				messageCacheMock.EXPECT().
					GetAllSentMessages(gomock.Any(), 10).
					Return(cachedMessages, nil).
					Times(1)

				messageRepoMock.EXPECT().
					GetSentMessages(7).
					Return([]models.Message{}, nil).
					Times(1)

				service := services.NewMessageService(
					messageRepoMock,
					messageCacheMock,
					nil,
					logger,
				)

				app := fiber.New()
				fiberCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				defer app.ReleaseCtx(fiberCtx)

				responses, err := service.ListSentMessages(fiberCtx, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(responses).To(HaveLen(3))
				// Should be sorted by sentAt descending
				Expect(responses[0].MessageID).To(Equal(int64(1))) // newest (-1h)
				Expect(responses[1].MessageID).To(Equal(int64(2))) // middle (-2h)
				Expect(responses[2].MessageID).To(Equal(int64(3))) // oldest (-3h)
			})

			It("should sort combined cache and DB messages by sentAt descending", func() {
				now := time.Now()
				// Cache has older message
				cachedMessages := []models.SentMessageCache{
					{
						MessageID:         1,
						ExternalMessageID: "cache-ext-1",
						To:                "+905551111111",
						Content:           "Cached Old Message",
						SentAt:            now.Add(-3 * time.Hour), // older
					},
				}

				// DB has newer messages
				dbMessages := []models.Message{
					{
						ID:                2,
						To:                "+905552222222",
						Content:           "DB Newest Message",
						Status:            models.StatusSent,
						ExternalMessageID: "db-ext-2",
						SentAt:            now.Add(-1 * time.Hour), // newest
					},
					{
						ID:                3,
						To:                "+905553333333",
						Content:           "DB Middle Message",
						Status:            models.StatusSent,
						ExternalMessageID: "db-ext-3",
						SentAt:            now.Add(-2 * time.Hour), // middle
					},
				}

				messageCacheMock.EXPECT().
					GetAllSentMessages(gomock.Any(), 10).
					Return(cachedMessages, nil).
					Times(1)

				messageRepoMock.EXPECT().
					GetSentMessages(9).
					Return(dbMessages, nil).
					Times(1)

				service := services.NewMessageService(
					messageRepoMock,
					messageCacheMock,
					nil,
					logger,
				)

				app := fiber.New()
				fiberCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				defer app.ReleaseCtx(fiberCtx)

				responses, err := service.ListSentMessages(fiberCtx, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(responses).To(HaveLen(3))
				// Should be sorted by sentAt descending regardless of source
				Expect(responses[0].MessageID).To(Equal(int64(2))) // DB newest (-1h)
				Expect(responses[1].MessageID).To(Equal(int64(3))) // DB middle (-2h)
				Expect(responses[2].MessageID).To(Equal(int64(1))) // Cache oldest (-3h)
			})

			It("should fall back to database when cache is empty", func() {
				messageCacheMock.EXPECT().
					GetAllSentMessages(gomock.Any(), 10).
					Return([]models.SentMessageCache{}, nil).
					Times(1)

				extID := "db-ext-1"
				sentAt := time.Now()
				dbMessages := []models.Message{
					{
						ID:                1,
						To:                "+905551111111",
						Content:           "DB Message 1",
						Status:            models.StatusSent,
						ExternalMessageID: extID,
						SentAt:            sentAt,
					},
				}

				messageRepoMock.EXPECT().
					GetSentMessages(10).
					Return(dbMessages, nil).
					Times(1)

				service := services.NewMessageService(
					messageRepoMock,
					messageCacheMock,
					nil,
					logger,
				)

				app := fiber.New()
				fiberCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				defer app.ReleaseCtx(fiberCtx)

				responses, err := service.ListSentMessages(fiberCtx, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(responses).To(HaveLen(1))
				Expect(responses[0].ExternalMessageID).To(Equal("db-ext-1"))
			})

			It("should fall back to database when cache fails", func() {
				messageCacheMock.EXPECT().
					GetAllSentMessages(gomock.Any(), 10).
					Return(nil, errors.New("cache error")).
					Times(1)

				messageRepoMock.EXPECT().
					GetSentMessages(10).
					Return([]models.Message{}, nil).
					Times(1)

				service := services.NewMessageService(
					messageRepoMock,
					messageCacheMock,
					nil,
					logger,
				)

				app := fiber.New()
				fiberCtx := app.AcquireCtx(&fasthttp.RequestCtx{})
				defer app.ReleaseCtx(fiberCtx)

				responses, err := service.ListSentMessages(fiberCtx, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(responses).To(BeEmpty())
			})
		})
	})

	Describe("Integration: Full Message Flow", func() {
		Context("when a message goes through the entire lifecycle", func() {
			It("should correctly transition from pending to sent to cached", func() {
				msg, err := messageRepository.CreateMessage("+905559999999", "Lifecycle Test Message")
				Expect(err).NotTo(HaveOccurred())
				Expect(msg).NotTo(BeNil())
				Expect(msg.Status).To(Equal(models.StatusPending))

				unsentMsgs, err := messageRepository.GetUnsentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(unsentMsgs).To(HaveLen(1))
				Expect(unsentMsgs[0].ID).To(Equal(msg.ID))

				externalID := "lifecycle-ext-001"
				sentAt := time.Now()
				err = messageRepository.UpdateMessageStatus(msg.ID, models.StatusSent, &externalID, &sentAt)
				Expect(err).NotTo(HaveOccurred())

				unsentMsgs, err = messageRepository.GetUnsentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(unsentMsgs).To(BeEmpty())

				sentMsgs, err := messageRepository.GetSentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(sentMsgs).To(HaveLen(1))
				Expect(sentMsgs[0].ID).To(Equal(msg.ID))
				Expect(sentMsgs[0].Status).To(Equal(models.StatusSent))
				Expect(sentMsgs[0].ExternalMessageID).To(Equal(externalID))

				cacheData := models.SentMessageCache{
					MessageID:         msg.ID,
					ExternalMessageID: externalID,
					To:                msg.To,
					Content:           msg.Content,
					SentAt:            sentAt,
				}
				err = messageCacheRepository.CacheSentMessage(ctx, cacheData)
				Expect(err).NotTo(HaveOccurred())

				cachedMsgs, err := messageCacheRepository.GetAllSentMessages(ctx, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(cachedMsgs).To(HaveLen(1))
				Expect(cachedMsgs[0].MessageID).To(Equal(msg.ID))
				Expect(cachedMsgs[0].ExternalMessageID).To(Equal(externalID))
				Expect(cachedMsgs[0].To).To(Equal(msg.To))
				Expect(cachedMsgs[0].Content).To(Equal(msg.Content))
			})
		})

		Context("when multiple messages are processed", func() {
			It("should handle batch processing correctly", func() {
				messages := make([]*models.Message, 5)
				for i := 0; i < 5; i++ {
					msg, err := messageRepository.CreateMessage(
						"+90555000000"+string(rune('0'+i)),
						"Batch Message "+string(rune('A'+i)),
					)
					Expect(err).NotTo(HaveOccurred())
					messages[i] = msg
				}

				unsentMsgs, err := messageRepository.GetUnsentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(unsentMsgs).To(HaveLen(5))

				for i, msg := range messages {
					externalID := "batch-ext-" + string(rune('0'+i))
					sentAt := time.Now()
					err = messageRepository.UpdateMessageStatus(msg.ID, models.StatusSent, &externalID, &sentAt)
					Expect(err).NotTo(HaveOccurred())

					cacheData := models.SentMessageCache{
						MessageID:         msg.ID,
						ExternalMessageID: externalID,
						To:                msg.To,
						Content:           msg.Content,
						SentAt:            sentAt,
					}
					err = messageCacheRepository.CacheSentMessage(ctx, cacheData)
					Expect(err).NotTo(HaveOccurred())
				}

				unsentMsgs, err = messageRepository.GetUnsentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(unsentMsgs).To(BeEmpty())

				sentMsgs, err := messageRepository.GetSentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(sentMsgs).To(HaveLen(5))

				cachedMsgs, err := messageCacheRepository.GetAllSentMessages(ctx, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(cachedMsgs).To(HaveLen(5))
			})
		})
	})
})
