package repository

import (
	"database/sql"
	"esbio/api/internal/domain"
	"fmt"
)

type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByID(userID int) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	UpdateUser(user *domain.User) error
	DeleteUser(userID int) error
	DeleteUserAndData(userID int) error
	UpdatePassword(userID int, hashedPassword string) error
	SetVerificationToken(userID int, token string, expires string) error
	VerifyEmail(token string) (*domain.User, error)
	SetResetToken(email string, token string, expires string) error
	GetUserByResetToken(token string) (*domain.User, error)
	ClearResetToken(userID int) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *domain.User) error {
	query := `
		INSERT INTO users (first_name, last_name, company_name, org_number, email, password_hash, role, email_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING user_id
	`
	err := r.db.QueryRow(query, user.FirstName, user.LastName, user.CompanyName, user.OrgNumber, user.Email, user.PasswordHash, user.Role, user.EmailVerified).Scan(&user.UserID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	user.Name = user.FirstName + " " + user.LastName
	return nil
}

func scanUser(row interface{ Scan(...interface{}) error }) (*domain.User, error) {
	user := &domain.User{}
	var companyName, orgNumber sql.NullString
	err := row.Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&companyName,
		&orgNumber,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.EmailVerified,
	)
	if err != nil {
		return nil, err
	}
	user.Name = user.FirstName + " " + user.LastName
	user.CompanyName = companyName.String
	user.OrgNumber = orgNumber.String
	return user, nil
}

func (r *userRepository) GetUserByID(userID int) (*domain.User, error) {
	query := `
		SELECT user_id, first_name, last_name, company_name, org_number, email, password_hash, role, email_verified
		FROM users
		WHERE user_id = $1
	`
	user, err := scanUser(r.db.QueryRow(query, userID))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *userRepository) GetUserByEmail(email string) (*domain.User, error) {
	query := `
		SELECT user_id, first_name, last_name, company_name, org_number, email, password_hash, role, email_verified
		FROM users
		WHERE email = $1
	`
	user, err := scanUser(r.db.QueryRow(query, email))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *userRepository) UpdateUser(user *domain.User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, company_name = $3, org_number = $4, email = $5, password_hash = $6, role = $7
		WHERE user_id = $8
	`
	_, err := r.db.Exec(query, user.FirstName, user.LastName, user.CompanyName, user.OrgNumber, user.Email, user.PasswordHash, user.Role, user.UserID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	user.Name = user.FirstName + " " + user.LastName
	return nil
}

func (r *userRepository) DeleteUser(userID int) error {
	query := `DELETE FROM users WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// DeleteUserAndData deletes a user and all their associated data in a transaction
func (r *userRepository) DeleteUserAndData(userID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete in dependency order for all companies owned by this user
	// 1. Line items (depend on vouchers)
	_, err = tx.Exec(`
		DELETE FROM line_items WHERE voucher_id IN (
			SELECT voucher_id FROM vouchers WHERE company_id IN (
				SELECT company_id FROM companies WHERE created_by = $1
			)
		)
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete line items: %w", err)
	}

	// 2. Vouchers (depend on companies)
	_, err = tx.Exec(`
		DELETE FROM vouchers WHERE company_id IN (
			SELECT company_id FROM companies WHERE created_by = $1
		)
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete vouchers: %w", err)
	}

	// 3-6. Bank data and categorization rules (tables may not exist in all environments)
	optionalTables := []string{"bank_transactions", "bank_accounts", "bank_connections", "bank_categorization_rules"}
	for _, table := range optionalTables {
		var exists bool
		err = tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)`, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}
		if exists {
			_, err = tx.Exec(fmt.Sprintf(`
				DELETE FROM %s WHERE company_id IN (
					SELECT company_id FROM companies WHERE created_by = $1
				)
			`, table), userID)
			if err != nil {
				return fmt.Errorf("failed to delete %s: %w", table, err)
			}
		}
	}

	// 7. Scheduled tasks
	_, err = tx.Exec(`
		DELETE FROM scheduled_tasks WHERE company_id IN (
			SELECT company_id FROM companies WHERE created_by = $1
		)
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete scheduled tasks: %w", err)
	}

	// 8. Agent messages and usage (cascade from user)
	_, err = tx.Exec(`DELETE FROM agent_messages WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete agent messages: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM agent_usage WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete agent usage: %w", err)
	}

	// 9. Companies
	_, err = tx.Exec(`DELETE FROM companies WHERE created_by = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete companies: %w", err)
	}

	// 10. User
	_, err = tx.Exec(`DELETE FROM users WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return tx.Commit()
}

func (r *userRepository) UpdatePassword(userID int, hashedPassword string) error {
	_, err := r.db.Exec(`UPDATE users SET password_hash = $1 WHERE user_id = $2`, hashedPassword, userID)
	return err
}

func (r *userRepository) SetVerificationToken(userID int, token string, expires string) error {
	_, err := r.db.Exec(
		`UPDATE users SET verification_token = $1, verification_token_expires = $2 WHERE user_id = $3`,
		token, expires, userID,
	)
	return err
}

func (r *userRepository) VerifyEmail(token string) (*domain.User, error) {
	query := `
		UPDATE users
		SET email_verified = TRUE, verification_token = NULL, verification_token_expires = NULL
		WHERE verification_token = $1 AND verification_token_expires > NOW()
		RETURNING user_id, first_name, last_name, company_name, org_number, email, password_hash, role, email_verified
	`
	return scanUser(r.db.QueryRow(query, token))
}

func (r *userRepository) SetResetToken(email string, token string, expires string) error {
	result, err := r.db.Exec(
		`UPDATE users SET reset_token = $1, reset_token_expires = $2 WHERE email = $3`,
		token, expires, email,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *userRepository) GetUserByResetToken(token string) (*domain.User, error) {
	query := `
		SELECT user_id, first_name, last_name, company_name, org_number, email, password_hash, role, email_verified
		FROM users
		WHERE reset_token = $1 AND reset_token_expires > NOW()
	`
	user, err := scanUser(r.db.QueryRow(query, token))
	if err != nil {
		return nil, fmt.Errorf("invalid or expired reset token")
	}
	return user, nil
}

func (r *userRepository) ClearResetToken(userID int) error {
	_, err := r.db.Exec(
		`UPDATE users SET reset_token = NULL, reset_token_expires = NULL WHERE user_id = $1`,
		userID,
	)
	return err
}
