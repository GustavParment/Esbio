package repository

import (
	"esbio/api/internal/domain"
	"database/sql"
	"fmt"
)

type BankConnectionRepository interface {
	Create(conn *domain.BankConnection) error
	GetByID(connectionID int) (*domain.BankConnection, error)
	GetByCompanyID(companyID int) ([]*domain.BankConnection, error)
	UpdateStatus(connectionID int, status string) error
	UpdateLastSynced(connectionID int) error
	UpdateAccessToken(connectionID int, encryptedToken string) error
	Delete(connectionID int) error
	GetExpiringConnections(daysAhead int) ([]*domain.BankConnection, error)
	GetActiveConnections() ([]*domain.BankConnection, error)
}

type bankConnectionRepository struct {
	db *sql.DB
}

func NewBankConnectionRepository(db *sql.DB) BankConnectionRepository {
	return &bankConnectionRepository{db: db}
}

func (r *bankConnectionRepository) Create(conn *domain.BankConnection) error {
	query := `
		INSERT INTO bank_connections (company_id, created_by, tink_user_id, tink_credentials_id, bank_name, status, access_token_encrypted, consent_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING connection_id, created_at, updated_at
	`
	return r.db.QueryRow(query,
		conn.CompanyID,
		conn.CreatedBy,
		conn.TinkUserID,
		conn.TinkCredentialsID,
		conn.BankName,
		conn.Status,
		conn.AccessTokenEncrypted,
		conn.ConsentExpiresAt,
	).Scan(&conn.ConnectionID, &conn.CreatedAt, &conn.UpdatedAt)
}

func (r *bankConnectionRepository) GetByID(connectionID int) (*domain.BankConnection, error) {
	query := `
		SELECT connection_id, company_id, created_by, tink_user_id, tink_credentials_id, bank_name, status, access_token_encrypted, consent_expires_at, last_synced_at, created_at, updated_at
		FROM bank_connections WHERE connection_id = $1
	`
	conn := &domain.BankConnection{}
	err := r.db.QueryRow(query, connectionID).Scan(
		&conn.ConnectionID, &conn.CompanyID, &conn.CreatedBy,
		&conn.TinkUserID, &conn.TinkCredentialsID, &conn.BankName,
		&conn.Status, &conn.AccessTokenEncrypted, &conn.ConsentExpiresAt,
		&conn.LastSyncedAt, &conn.CreatedAt, &conn.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bank connection not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bank connection: %w", err)
	}
	return conn, nil
}

func (r *bankConnectionRepository) GetByCompanyID(companyID int) ([]*domain.BankConnection, error) {
	query := `
		SELECT connection_id, company_id, created_by, tink_user_id, tink_credentials_id, bank_name, status, consent_expires_at, last_synced_at, created_at, updated_at
		FROM bank_connections WHERE company_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank connections: %w", err)
	}
	defer rows.Close()

	var connections []*domain.BankConnection
	for rows.Next() {
		conn := &domain.BankConnection{}
		err := rows.Scan(
			&conn.ConnectionID, &conn.CompanyID, &conn.CreatedBy,
			&conn.TinkUserID, &conn.TinkCredentialsID, &conn.BankName,
			&conn.Status, &conn.ConsentExpiresAt,
			&conn.LastSyncedAt, &conn.CreatedAt, &conn.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bank connection: %w", err)
		}
		connections = append(connections, conn)
	}
	return connections, nil
}

func (r *bankConnectionRepository) UpdateStatus(connectionID int, status string) error {
	query := `UPDATE bank_connections SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE connection_id = $2`
	_, err := r.db.Exec(query, status, connectionID)
	return err
}

func (r *bankConnectionRepository) UpdateLastSynced(connectionID int) error {
	query := `UPDATE bank_connections SET last_synced_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE connection_id = $1`
	_, err := r.db.Exec(query, connectionID)
	return err
}

func (r *bankConnectionRepository) UpdateAccessToken(connectionID int, encryptedToken string) error {
	query := `UPDATE bank_connections SET access_token_encrypted = $1, updated_at = CURRENT_TIMESTAMP WHERE connection_id = $2`
	_, err := r.db.Exec(query, encryptedToken, connectionID)
	return err
}

func (r *bankConnectionRepository) Delete(connectionID int) error {
	query := `DELETE FROM bank_connections WHERE connection_id = $1`
	_, err := r.db.Exec(query, connectionID)
	return err
}

func (r *bankConnectionRepository) GetExpiringConnections(daysAhead int) ([]*domain.BankConnection, error) {
	query := `
		SELECT connection_id, company_id, created_by, tink_user_id, tink_credentials_id, bank_name, status, consent_expires_at, last_synced_at, created_at, updated_at
		FROM bank_connections
		WHERE status = 'active' AND consent_expires_at < CURRENT_TIMESTAMP + MAKE_INTERVAL(days => $1)
		ORDER BY consent_expires_at
	`
	rows, err := r.db.Query(query, daysAhead)
	if err != nil {
		return nil, fmt.Errorf("failed to get expiring connections: %w", err)
	}
	defer rows.Close()

	var connections []*domain.BankConnection
	for rows.Next() {
		conn := &domain.BankConnection{}
		err := rows.Scan(
			&conn.ConnectionID, &conn.CompanyID, &conn.CreatedBy,
			&conn.TinkUserID, &conn.TinkCredentialsID, &conn.BankName,
			&conn.Status, &conn.ConsentExpiresAt,
			&conn.LastSyncedAt, &conn.CreatedAt, &conn.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bank connection: %w", err)
		}
		connections = append(connections, conn)
	}
	return connections, nil
}

func (r *bankConnectionRepository) GetActiveConnections() ([]*domain.BankConnection, error) {
	query := `
		SELECT connection_id, company_id, created_by, tink_user_id, tink_credentials_id, bank_name, status, access_token_encrypted, consent_expires_at, last_synced_at, created_at, updated_at
		FROM bank_connections WHERE status = 'active'
		ORDER BY last_synced_at NULLS FIRST
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active connections: %w", err)
	}
	defer rows.Close()

	var connections []*domain.BankConnection
	for rows.Next() {
		conn := &domain.BankConnection{}
		err := rows.Scan(
			&conn.ConnectionID, &conn.CompanyID, &conn.CreatedBy,
			&conn.TinkUserID, &conn.TinkCredentialsID, &conn.BankName,
			&conn.Status, &conn.AccessTokenEncrypted, &conn.ConsentExpiresAt,
			&conn.LastSyncedAt, &conn.CreatedAt, &conn.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bank connection: %w", err)
		}
		connections = append(connections, conn)
	}
	return connections, nil
}
