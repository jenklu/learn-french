package services

import (
	"context"
	"fmt"

	"github.com/jenklu/learn-french/internal/database"
	"github.com/jenklu/learn-french/internal/models"
)

type AdminService struct {
	db *database.Database
}

func NewAdminService(db *database.Database) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.db.GetAllUsers(ctx)
}

func (s *AdminService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	return s.db.GetUserByID(ctx, id)
}

func (s *AdminService) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if err := s.db.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (s *AdminService) UpdateUser(ctx context.Context, id int, username, email, timezone string, isAdmin bool) (*models.User, error) {
	if err := s.db.UpdateUser(ctx, id, username, email, timezone, isAdmin); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return s.db.GetUserByID(ctx, id)
}

func (s *AdminService) DeleteUser(ctx context.Context, id int) error {
	return s.db.DeleteUser(ctx, id)
}

func (s *AdminService) UpdateUserPassword(ctx context.Context, id int, password string) error {
	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.db.UpdateUserPassword(ctx, id, string(hashedPassword))
}

func hashPassword(password string) (string, error) {
	// Import bcrypt here to avoid circular dependency
	bytes, err := bcryptHash([]byte(password))
	return string(bytes), err
}

// This is a workaround for the bcrypt import issue
import "golang.org/x/crypto/bcrypt"

func bcryptHash(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}
