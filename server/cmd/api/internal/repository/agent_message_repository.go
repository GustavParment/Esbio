package repository

import (
	"esbio/api/internal/domain"
	"database/sql"
	"fmt"
)

type AgentMessageRepository interface {
	SaveMessage(msg *domain.AgentMessage) error
	GetMessagesByConversation(conversationID string) ([]*domain.AgentMessage, error)
	GetConversationsByCompanyID(companyID int) ([]string, error)
}

type agentMessageRepository struct {
	db *sql.DB
}

func NewAgentMessageRepository(db *sql.DB) AgentMessageRepository {
	return &agentMessageRepository{db: db}
}

func (r *agentMessageRepository) SaveMessage(msg *domain.AgentMessage) error {
	query := `
		INSERT INTO agent_messages (user_id, company_id, conversation_id, role, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING message_id, created_at
	`
	err := r.db.QueryRow(query, msg.UserID, msg.CompanyID, msg.ConversationID, msg.Role, msg.Content).
		Scan(&msg.MessageID, &msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save agent message: %w", err)
	}
	return nil
}

func (r *agentMessageRepository) GetMessagesByConversation(conversationID string) ([]*domain.AgentMessage, error) {
	query := `
		SELECT message_id, user_id, company_id, conversation_id, role, content, created_at
		FROM agent_messages WHERE conversation_id = $1 ORDER BY created_at
	`
	rows, err := r.db.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.AgentMessage
	for rows.Next() {
		msg := &domain.AgentMessage{}
		err := rows.Scan(&msg.MessageID, &msg.UserID, &msg.CompanyID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *agentMessageRepository) GetConversationsByCompanyID(companyID int) ([]string, error) {
	query := `
		SELECT DISTINCT conversation_id FROM agent_messages
		WHERE company_id = $1 ORDER BY conversation_id DESC
	`
	rows, err := r.db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan conversation id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
