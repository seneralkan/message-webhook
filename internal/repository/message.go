package repository

import (
	"database/sql"
	"fmt"
	"time"

	"go-template-microservice/internal/models"
	"go-template-microservice/pkg/sqlite"

	"github.com/sirupsen/logrus"
)

type MessageRepository interface {
	// GetUnsentMessages retrieves messages with PENDING status, limited by the given count
	GetUnsentMessages(limit int) ([]models.Message, error)
	// UpdateMessageStatus updates the status of a message and optionally sets external message ID and sent time
	UpdateMessageStatus(messageID int64, status models.Status, externalMessageID *string, sentAt *time.Time) error
	// CreateMessage creates a new message record in the database
	CreateMessage(to, content string) (*models.Message, error)
	// GetSentMessages retrieves messages with SENT status, limited by the given count and ordered by sent_at descending
	GetSentMessages(limit int) ([]models.Message, error)
}

type messageRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewMessageRepository(sqlite sqlite.ISqliteInstance, logger *logrus.Logger) MessageRepository {
	return &messageRepository{
		db:     sqlite.Database(),
		logger: logger,
	}
}

func (r *messageRepository) GetUnsentMessages(limit int) ([]models.Message, error) {
	query := `
		SELECT id, "to", content, status, external_message_id, sent_at, created_at, updated_at
		FROM messages
		WHERE status = ?
		ORDER BY created_at ASC
		LIMIT ?
	`

	rows, err := r.db.Query(query, models.StatusPending, limit)
	if err != nil {
		r.logger.WithError(err).Error("Failed to query unsent messages")
		return nil, fmt.Errorf("failed to query unsent messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var sentAt sql.NullTime
		err := rows.Scan(
			&msg.ID,
			&msg.To,
			&msg.Content,
			&msg.Status,
			&msg.ExternalMessageID,
			&sentAt,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan message row")
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}
		if sentAt.Valid {
			msg.SentAt = sentAt.Time
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating message rows")
		return nil, fmt.Errorf("error iterating message rows: %w", err)
	}

	r.logger.WithField("count", len(messages)).Debug("Retrieved unsent messages")
	return messages, nil
}

func (r *messageRepository) UpdateMessageStatus(messageID int64, status models.Status, externalMessageID *string, sentAt *time.Time) error {
	query := `
		UPDATE messages
		SET status = ?, external_message_id = ?, sent_at = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	extID := ""
	if externalMessageID != nil {
		extID = *externalMessageID
	}
	result, err := r.db.Exec(query, status, extID, sentAt, now, messageID)
	if err != nil {
		r.logger.WithError(err).WithField("messageID", messageID).Error("Failed to update message status")
		return fmt.Errorf("failed to update message status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.WithField("messageID", messageID).Warn("No message found with given ID")
		return fmt.Errorf("no message found with ID: %d", messageID)
	}

	r.logger.WithFields(logrus.Fields{
		"messageID": messageID,
		"status":    status,
	}).Debug("Message status updated successfully")

	return nil
}

// CreateMessage creates a new message with PENDING status
func (r *messageRepository) CreateMessage(to, content string) (*models.Message, error) {
	if len(content) > 160 {
		return nil, fmt.Errorf("content exceeds 160 character limit")
	}

	query := `
		INSERT INTO messages ("to", content, status, external_message_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(query, to, content, models.StatusPending, "", now, now)
	if err != nil {
		r.logger.WithError(err).Error("Failed to create message")
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logger.WithError(err).Error("Failed to get last insert ID")
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	message := &models.Message{
		ID:                id,
		To:                to,
		Content:           content,
		Status:            models.StatusPending,
		ExternalMessageID: "",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	r.logger.WithField("messageID", id).Debug("Message created successfully")
	return message, nil
}

func (r *messageRepository) GetSentMessages(limit int) ([]models.Message, error) {
	query := `
		SELECT id, "to", content, status, external_message_id, sent_at, created_at, updated_at
		FROM messages
		WHERE status = ?
		ORDER BY sent_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, models.StatusSent, limit)
	if err != nil {
		r.logger.WithError(err).Error("Failed to query sent messages")
		return nil, fmt.Errorf("failed to query sent messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.To,
			&msg.Content,
			&msg.Status,
			&msg.ExternalMessageID,
			&msg.SentAt,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan message row")
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating message rows")
		return nil, fmt.Errorf("error iterating message rows: %w", err)
	}

	r.logger.WithField("count", len(messages)).Debug("Retrieved sent messages")
	return messages, nil
}
