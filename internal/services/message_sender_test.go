package services_test

import (
	"context"
	"net/http"

	"go-template-microservice/internal/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessageSenderService", func() {
	Describe("Send", func() {
		Context("when webhook returns 202 Accepted", func() {
			It("should successfully send the message and return response", func() {
				server := createMockWebhookServer(http.StatusAccepted, `{"message":"Accepted","messageId":"ext-12345"}`)
				defer server.Close()

				sender := services.NewMessageSenderService(server.URL, "test-auth-key", logger)
				resp, err := sender.Send(context.Background(), "+905551234567", "Hello World")

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.MessageID).To(Equal("ext-12345"))
				Expect(resp.Message).To(Equal("Accepted"))
			})
		})

		Context("when webhook returns non-202 status", func() {
			It("should return an error for 400 Bad Request", func() {
				server := createMockWebhookServer(http.StatusBadRequest, `{"error":"bad request"}`)
				defer server.Close()

				sender := services.NewMessageSenderService(server.URL, "test-auth-key", logger)
				resp, err := sender.Send(context.Background(), "+905551234567", "Hello World")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("status code: 400"))
				Expect(resp).To(BeNil())
			})

			It("should return an error for 500 Internal Server Error", func() {
				server := createMockWebhookServer(http.StatusInternalServerError, `{"error":"internal error"}`)
				defer server.Close()

				sender := services.NewMessageSenderService(server.URL, "test-auth-key", logger)
				resp, err := sender.Send(context.Background(), "+905551234567", "Hello World")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("status code: 500"))
				Expect(resp).To(BeNil())
			})
		})

		Context("when webhook server is unreachable", func() {
			It("should return an error", func() {
				sender := services.NewMessageSenderService("http://localhost:99999", "test-auth-key", logger)
				resp, err := sender.Send(context.Background(), "+905551234567", "Hello World")

				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
			})
		})

		Context("when webhook returns invalid JSON", func() {
			It("should return a decode error", func() {
				server := createMockWebhookServer(http.StatusAccepted, `invalid json`)
				defer server.Close()

				sender := services.NewMessageSenderService(server.URL, "test-auth-key", logger)
				resp, err := sender.Send(context.Background(), "+905551234567", "Hello World")

				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
			})
		})
	})
})
