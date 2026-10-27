package services

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/jenklu/learn-french/internal/database"
	"github.com/jenklu/learn-french/internal/models"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrUsernameExists   = errors.New("username already exists")
	ErrEmailExists      = errors.New("email already exists")
)

type AuthService struct {
	db *database.Database
}

func NewAuthService(db *database.Database) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(ctx context.Context, req *models.UserRegistration) (*models.User, error) {
	// Check if username exists
	if _, err := s.db.GetUserByUsername(ctx, req.Username); err == nil {
		return nil, ErrUsernameExists
	}

	// Check if email exists
	if _, err := s.db.GetUserByEmail(ctx, req.Email); err == nil {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Timezone:     req.Timezone,
		IsAdmin:      false,
	}

	if err := s.db.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default reminder (disabled)
	if err := s.db.UpdateReminder(ctx, user.ID, false); err != nil {
		return nil, fmt.Errorf("failed to create reminder: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *models.UserLogin) (*models.User, error) {
	user, err := s.db.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

func (s *AuthService) UpdateTimezone(ctx context.Context, userID int, timezone string) error {
	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	return s.db.UpdateUser(ctx, userID, user.Username, user.Email, timezone, user.IsAdmin)
}

func (s *AuthService) UpdatePassword(ctx context.Context, userID int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.db.UpdateUserPassword(ctx, userID, string(hashedPassword))
}
