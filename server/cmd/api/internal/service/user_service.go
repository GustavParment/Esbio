/*
UserService provides business logic for user management.
It interacts with the UserRepository for data persistence and uses
the validator package for input validation.
*/
package service


import (
	"esbio/api/internal/domain"
	"esbio/api/internal/repository"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repository repository.UserRepository
	validate   *validator.Validate
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repository: repo,
		validate:   validator.New(),
	}
}

func (s *UserService) CreateUser(user *domain.User) error {
	// 1. Validate the input using validator
	if err := s.validate.Struct(user); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 2. Check if user already exists
	existingUser, err := s.repository.GetUserByEmail(user.Email)
	if err == nil && existingUser != nil {
		return errors.New("user with this email already exists")
	}

	// 3. Hash the password (assuming PasswordHash contains plain text password at this point)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)

	// 4. Set default role if not provided
	if user.Role == "" {
		user.Role = "Bookkeeper"
	}

	// 5. Save to database via repository
	err = s.repository.CreateUser(user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(userID int) (*domain.User, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.repository.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*domain.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	user, err := s.repository.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *UserService) UpdateUser(user *domain.User) error {
	// 1. Validate input using validator
	if err := s.validate.Struct(user); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 2. Check if user exists
	existingUser, err := s.repository.GetUserByID(user.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 3. If email is being changed, check if new email already exists
	if existingUser.Email != user.Email {
		userWithEmail, err := s.repository.GetUserByEmail(user.Email)
		if err == nil && userWithEmail != nil {
			return errors.New("email already exists")
		}
	}

	// 4. If password is being updated (non-empty PasswordHash), hash it
	if user.PasswordHash != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	} else {
		// Keep existing password if no new password provided
		user.PasswordHash = existingUser.PasswordHash
	}

	// 5. Update in database
	err = s.repository.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *UserService) DeleteUser(userID int) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	err := s.repository.DeleteUser(userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// DeleteUserAndData deletes a user and all their associated data
func (s *UserService) DeleteUserAndData(userID int) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	return s.repository.DeleteUserAndData(userID)
}
