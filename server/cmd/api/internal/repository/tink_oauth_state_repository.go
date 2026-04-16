package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type TinkOAuthStateRepository interface {
	Create(stateID string, companyID, userID int, expiresAt time.Time) error
	GetAndDelete(stateID string) (companyID, userID int, err error)
	CleanupExpired() error
}

type tinkOAuthStateRepository struct {
	db *sql.DB
}

func NewTinkOAuthStateRepository(db *sql.DB) TinkOAuthStateRepository {
	return &tinkOAuthStateRepository{db: db}
}

func (r *tinkOAuthStateRepository) Create(stateID string, companyID, userID int, expiresAt time.Time) error {
	query := `INSERT INTO tink_oauth_state (state_id, company_id, user_id, expires_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, stateID, companyID, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create oauth state: %w", err)
	}
	return nil
}

// GetAndDelete atomically retrieves and deletes the state row, preventing replay attacks.
// Returns an error if the state is expired or not found.
func (r *tinkOAuthStateRepository) GetAndDelete(stateID string) (int, int, error) {
	query := `DELETE FROM tink_oauth_state WHERE state_id = $1 AND expires_at > CURRENT_TIMESTAMP RETURNING company_id, user_id`
	var companyID, userID int
	err := r.db.QueryRow(query, stateID).Scan(&companyID, &userID)
	if err == sql.ErrNoRows {
		return 0, 0, fmt.Errorf("invalid or expired oauth state")
	}
	if err != nil {
		return 0, 0, fmt.Errorf("failed to validate oauth state: %w", err)
	}
	return companyID, userID, nil
}

func (r *tinkOAuthStateRepository) CleanupExpired() error {
	query := `DELETE FROM tink_oauth_state WHERE expires_at < CURRENT_TIMESTAMP`
	_, err := r.db.Exec(query)
	return err
}
