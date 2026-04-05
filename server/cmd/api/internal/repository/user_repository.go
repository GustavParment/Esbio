package repository

import (
	"cmd/api/internal/domain"
	"database/sql"
	"fmt"
)

type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByID(userID int) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	UpdateUser(user *domain.User) error
	DeleteUser(userID int) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *domain.User) error {
	query := `
		INSERT INTO users (first_name, last_name, email, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id
	`
	err := r.db.QueryRow(query, user.FirstName, user.LastName, user.Email, user.PasswordHash, user.Role).Scan(&user.UserID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	user.Name = user.FirstName + " " + user.LastName
	return nil
}

func scanUser(row interface{ Scan(...interface{}) error }) (*domain.User, error) {
	user := &domain.User{}
	err := row.Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
	)
	if err != nil {
		return nil, err
	}
	user.Name = user.FirstName + " " + user.LastName
	return user, nil
}

func (r *userRepository) GetUserByID(userID int) (*domain.User, error) {
	query := `
		SELECT user_id, first_name, last_name, email, password_hash, role
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
		SELECT user_id, first_name, last_name, email, password_hash, role
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
		SET first_name = $1, last_name = $2, email = $3, password_hash = $4, role = $5
		WHERE user_id = $6
	`
	_, err := r.db.Exec(query, user.FirstName, user.LastName, user.Email, user.PasswordHash, user.Role, user.UserID)
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
