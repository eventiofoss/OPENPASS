package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const minPasswordLen = 8

var (
	// ErrInvalidInput is returned when required fields are missing.
	ErrInvalidInput = errors.New("invalid input")
	// ErrDuplicateEmail means the email is already registered.
	ErrDuplicateEmail = errors.New(
		"an account with this email already exists",
	)
	// ErrInvalidCredentials is returned on bad login attempts.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// OrganizerRepo defines the database operations used by AuthService.
type OrganizerRepo interface {
	Create(ctx context.Context, org *models.Organizer) error
	FindByEmail(ctx context.Context, email string) (*models.Organizer, error)
	FindByID(ctx context.Context, id string) (*models.Organizer, error)
}

// AuthService contains authentication business logic.
type AuthService struct {
	repo OrganizerRepo
}

// NewAuthService returns a service backed by the concrete repository.
func NewAuthService(
	repo *repository.OrganizerRepository,
) *AuthService {
	return &AuthService{repo: repo}
}

// NewAuthServiceWithRepo returns a service using any OrganizerRepo.
func NewAuthServiceWithRepo(repo OrganizerRepo) *AuthService {
	return &AuthService{repo: repo}
}

// Register creates a new organizer account.
func (s *AuthService) Register(
	ctx context.Context,
	name, email, password string,
) (*models.Organizer, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))

	if name == "" || email == "" || len(password) < minPasswordLen {
		return nil, fmt.Errorf(
			"%w: name, email, and password (min %d chars) required",
			ErrInvalidInput, minPasswordLen,
		)
	}

	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(password), 12,
	)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	organizer := &models.Organizer{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashed),
		Role:         models.OrganizerRoleOrganizer,
	}

	if err := s.repo.Create(ctx, organizer); err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return nil, ErrDuplicateEmail
		}

		return nil, fmt.Errorf("creating organizer: %w", err)
	}

	return organizer, nil
}

// Login authenticates an organizer and returns a signed JWT.
func (s *AuthService) Login(
	ctx context.Context,
	email, password string,
) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	organizer, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("finding organizer: %w", err)
	}

	if organizer == nil {
		return "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(organizer.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := middleware.GenerateToken(
		*organizer, middleware.DefaultTokenTTL,
	)
	if err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}

	return token, nil
}

// GetOrganizer returns an organizer by ID for session resolution.
func (s *AuthService) GetOrganizer(
	ctx context.Context,
	id string,
) (*models.Organizer, error) {
	organizer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding organizer: %w", err)
	}

	if organizer == nil {
		return nil, ErrInvalidCredentials
	}

	return organizer, nil
}
