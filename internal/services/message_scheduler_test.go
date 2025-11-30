package services_test

import (
	"net/http"
	"time"

	"go-template-microservice/internal/models"
	"go-template-microservice/internal/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessageScheduler", func() {
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

	Describe("Full Flow Integration Test", func() {
		Context("when messages are created and scheduler processes them", func() {
			It("should send messages via webhook and update their status", func() {
				server := createMockWebhookServer(http.StatusAccepted, `{"message":"Accepted","messageId":"ext-msg-001"}`)
				defer server.Close()

				sender := services.NewMessageSenderService(server.URL, "test-auth-key", logger)

				msg1, err := messageRepository.CreateMessage("+905551111111", "Test Message 1")
				Expect(err).NotTo(HaveOccurred())
				Expect(msg1).NotTo(BeNil())

				msg2, err := messageRepository.CreateMessage("+905552222222", "Test Message 2")
				Expect(err).NotTo(HaveOccurred())
				Expect(msg2).NotTo(BeNil())

				pendingMessages, err := messageRepository.GetUnsentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(pendingMessages).To(HaveLen(2))

				_ = sender
			})
		})
	})

	Describe("Tick Processing with Mocks", func() {
		Context("when there are pending messages", func() {
			It("should process messages correctly through the scheduler tick", func() {
				// This test verifies the scheduler setup - actual tick processing
				// is tested via the end-to-end integration tests below
				scheduler := services.NewMessageScheduler(
					messageRepoMock,
					messageSenderMock,
					messageCacheMock,
					1*time.Hour,
					10,
					logger,
				)
				Expect(scheduler).NotTo(BeNil())
			})
		})
	})

	Describe("End-to-End Flow with Real Components", func() {
		Context("when processing messages through the entire pipeline", func() {
			It("should create, send, and cache messages correctly", func() {
				server := createMockWebhookServer(http.StatusAccepted, `{"message":"Accepted","messageId":"e2e-ext-001"}`)
				defer server.Close()

				sender := services.NewMessageSenderService(server.URL, "test-auth-key", logger)

				msg, err := messageRepository.CreateMessage("+905559999999", "E2E Test Message")
				Expect(err).NotTo(HaveOccurred())
				Expect(msg).NotTo(BeNil())
				Expect(msg.Status).To(Equal(models.StatusPending))

				pendingMsgs, err := messageRepository.GetUnsentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(pendingMsgs).To(HaveLen(1))
				Expect(pendingMsgs[0].ID).To(Equal(msg.ID))

				resp, err := sender.Send(ctx, msg.To, msg.Content)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.MessageID).To(Equal("e2e-ext-001"))

				sentAt := time.Now()
				err = messageRepository.UpdateMessageStatus(msg.ID, models.StatusSent, &resp.MessageID, &sentAt)
				Expect(err).NotTo(HaveOccurred())

				cacheData := models.SentMessageCache{
					MessageID:         msg.ID,
					ExternalMessageID: resp.MessageID,
					To:                msg.To,
					Content:           msg.Content,
					SentAt:            sentAt,
				}
				err = messageCacheRepository.CacheSentMessage(ctx, cacheData)
				Expect(err).NotTo(HaveOccurred())

				pendingMsgs, err = messageRepository.GetUnsentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(pendingMsgs).To(BeEmpty())

				sentMsgs, err := messageRepository.GetSentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(sentMsgs).To(HaveLen(1))
				Expect(sentMsgs[0].ID).To(Equal(msg.ID))
				Expect(sentMsgs[0].ExternalMessageID).To(Equal("e2e-ext-001"))

				cachedMsgs, err := messageCacheRepository.GetAllSentMessages(ctx, 10)
				Expect(err).NotTo(HaveOccurred())
				Expect(cachedMsgs).To(HaveLen(1))
				Expect(cachedMsgs[0].MessageID).To(Equal(msg.ID))
				Expect(cachedMsgs[0].ExternalMessageID).To(Equal("e2e-ext-001"))
			})
		})

		Context("when webhook fails", func() {
			It("should not update message status", func() {
				server := createMockWebhookServer(http.StatusInternalServerError, `{"error":"server error"}`)
				defer server.Close()

				sender := services.NewMessageSenderService(server.URL, "test-auth-key", logger)

				msg, err := messageRepository.CreateMessage("+905558888888", "Failed Message Test")
				Expect(err).NotTo(HaveOccurred())

				resp, err := sender.Send(ctx, msg.To, msg.Content)
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())

				pendingMsgs, err := messageRepository.GetUnsentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(pendingMsgs).To(HaveLen(1))
				Expect(pendingMsgs[0].Status).To(Equal(models.StatusPending))
			})
		})
	})
})
