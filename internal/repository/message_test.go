package repository_test

import (
	"time"

	"go-template-microservice/internal/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessageRepository", func() {
	var (
		createdMessageID int64
	)

	BeforeEach(func() {
		// Clean up any existing messages before each test
		_, err := mockSqlite.Database().Exec("DELETE FROM messages")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("CreateMessage", func() {
		Context("when creating a valid message", func() {
			It("should create the message successfully", func() {
				msg, err := messageRepository.CreateMessage("+905551234567", "Hello World")

				Expect(err).NotTo(HaveOccurred())
				Expect(msg).NotTo(BeNil())
				Expect(msg.ID).To(BeNumerically(">", 0))
				Expect(msg.To).To(Equal("+905551234567"))
				Expect(msg.Content).To(Equal("Hello World"))
				Expect(msg.Status).To(Equal(models.StatusPending))
				createdMessageID = msg.ID
			})
		})

		Context("when content exceeds 160 characters", func() {
			It("should return an error", func() {
				longContent := ""
				for i := 0; i < 161; i++ {
					longContent += "a"
				}

				msg, err := messageRepository.CreateMessage("+905551234567", longContent)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("content exceeds 160 character limit"))
				Expect(msg).To(BeNil())
			})
		})

		Context("when content is exactly 160 characters", func() {
			It("should create the message successfully", func() {
				content := ""
				for i := 0; i < 160; i++ {
					content += "a"
				}

				msg, err := messageRepository.CreateMessage("+905551234567", content)

				Expect(err).NotTo(HaveOccurred())
				Expect(msg).NotTo(BeNil())
				Expect(len(msg.Content)).To(Equal(160))
			})
		})
	})

	Describe("GetUnsentMessages", func() {
		BeforeEach(func() {
			// Create some test messages
			_, err := messageRepository.CreateMessage("+905551111111", "Message 1")
			Expect(err).NotTo(HaveOccurred())

			_, err = messageRepository.CreateMessage("+905552222222", "Message 2")
			Expect(err).NotTo(HaveOccurred())

			_, err = messageRepository.CreateMessage("+905553333333", "Message 3")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there are pending messages", func() {
			It("should return all pending messages up to the limit", func() {
				messages, err := messageRepository.GetUnsentMessages(10)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(HaveLen(3))
				for _, msg := range messages {
					Expect(msg.Status).To(Equal(models.StatusPending))
				}
			})

			It("should respect the limit parameter", func() {
				messages, err := messageRepository.GetUnsentMessages(2)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(HaveLen(2))
			})

			It("should return messages ordered by created_at ASC", func() {
				messages, err := messageRepository.GetUnsentMessages(10)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(HaveLen(3))
				Expect(messages[0].Content).To(Equal("Message 1"))
				Expect(messages[1].Content).To(Equal("Message 2"))
				Expect(messages[2].Content).To(Equal("Message 3"))
			})
		})

		Context("when there are no pending messages", func() {
			BeforeEach(func() {
				// Delete all messages
				_, err := mockSqlite.Database().Exec("DELETE FROM messages")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return an empty slice", func() {
				messages, err := messageRepository.GetUnsentMessages(10)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(BeEmpty())
			})
		})
	})

	Describe("UpdateMessageStatus", func() {
		BeforeEach(func() {
			msg, err := messageRepository.CreateMessage("+905551234567", "Test Message")
			Expect(err).NotTo(HaveOccurred())
			createdMessageID = msg.ID
		})

		Context("when updating to SENT status", func() {
			It("should update the message successfully", func() {
				externalID := "ext-123456"
				sentAt := time.Now()

				err := messageRepository.UpdateMessageStatus(createdMessageID, models.StatusSent, &externalID, &sentAt)

				Expect(err).NotTo(HaveOccurred())

				// Verify the update
				messages, err := messageRepository.GetUnsentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(BeEmpty()) // Should not find it as pending anymore
			})
		})

		Context("when updating to FAILED status", func() {
			It("should update the message successfully", func() {
				err := messageRepository.UpdateMessageStatus(createdMessageID, models.StatusFailed, nil, nil)

				Expect(err).NotTo(HaveOccurred())

				// Verify the update - should not be in pending anymore
				messages, err := messageRepository.GetUnsentMessages(10)
				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(BeEmpty())
			})
		})

		Context("when message ID does not exist", func() {
			It("should return an error", func() {
				nonExistentID := int64(99999)

				err := messageRepository.UpdateMessageStatus(nonExistentID, models.StatusSent, nil, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no message found with ID"))
			})
		})
	})

	Describe("GetSentMessages", func() {
		BeforeEach(func() {
			// Create and update messages to SENT status
			msg1, err := messageRepository.CreateMessage("+905551111111", "Sent Message 1")
			Expect(err).NotTo(HaveOccurred())
			extID1 := "ext-1"
			sentAt1 := time.Now().Add(-2 * time.Hour)
			err = messageRepository.UpdateMessageStatus(msg1.ID, models.StatusSent, &extID1, &sentAt1)
			Expect(err).NotTo(HaveOccurred())

			msg2, err := messageRepository.CreateMessage("+905552222222", "Sent Message 2")
			Expect(err).NotTo(HaveOccurred())
			extID2 := "ext-2"
			sentAt2 := time.Now().Add(-1 * time.Hour)
			err = messageRepository.UpdateMessageStatus(msg2.ID, models.StatusSent, &extID2, &sentAt2)
			Expect(err).NotTo(HaveOccurred())

			// Create a pending message (should not be returned)
			_, err = messageRepository.CreateMessage("+905553333333", "Pending Message")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there are sent messages", func() {
			It("should return only sent messages", func() {
				messages, err := messageRepository.GetSentMessages(10)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(HaveLen(2))
				for _, msg := range messages {
					Expect(msg.Status).To(Equal(models.StatusSent))
				}
			})

			It("should respect the limit parameter", func() {
				messages, err := messageRepository.GetSentMessages(1)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(HaveLen(1))
			})

			It("should return messages ordered by sent_at DESC", func() {
				messages, err := messageRepository.GetSentMessages(10)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(HaveLen(2))
				// Most recently sent should be first
				Expect(messages[0].Content).To(Equal("Sent Message 2"))
				Expect(messages[1].Content).To(Equal("Sent Message 1"))
			})
		})

		Context("when there are no sent messages", func() {
			BeforeEach(func() {
				// Delete all messages
				_, err := mockSqlite.Database().Exec("DELETE FROM messages")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return an empty slice", func() {
				messages, err := messageRepository.GetSentMessages(10)

				Expect(err).NotTo(HaveOccurred())
				Expect(messages).To(BeEmpty())
			})
		})
	})
})
