package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/eventiofoss/eventio/backend/internal/middleware"
	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/repository"
	"github.com/google/uuid"
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

// UserRepo defines the database operations used by AuthService.
type UserRepo interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
}

type userCreateAndClaimRepo interface {
	CreateAndClaim(ctx context.Context, user *models.User) error
}

// AuthService contains authentication business logic.
type AuthService struct {
	repo UserRepo
}

// NewAuthService returns a service backed by the concrete repository.
func NewAuthService(
	repo *repository.UserRepository,
) *AuthService {
	return &AuthService{repo: repo}
}

// NewAuthServiceWithRepo returns a service using any UserRepo.
func NewAuthServiceWithRepo(repo UserRepo) *AuthService {
	return &AuthService{repo: repo}
}

// Register creates a new user account.
func (s *AuthService) Register(
	ctx context.Context,
	name, email, password, role string,
) (*models.User, error) {
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

	// Account role must be organizer, volunteer, or participant.
	accountRole := models.UserRoleOrganizer
	role = strings.ToLower(strings.TrimSpace(role))
	switch role {
	case "", "organizer":
		accountRole = models.UserRoleOrganizer
	case "volunteer":
		accountRole = models.UserRoleVolunteer
	case "participant":
		accountRole = models.UserRoleParticipant
	default:
		return nil, fmt.Errorf(
			"%w: role must be organizer, volunteer, or participant",
			ErrInvalidInput,
		)
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashed),
		Role:         accountRole,
	}

	if createAndClaimRepo, ok := s.repo.(userCreateAndClaimRepo); ok {
		if err := createAndClaimRepo.CreateAndClaim(ctx, user); err != nil {
			if strings.Contains(err.Error(), "duplicate key value") {
				return nil, ErrDuplicateEmail
			}

			return nil, fmt.Errorf("creating user: %w", err)
		}

		return user, nil
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			return nil, ErrDuplicateEmail
		}

		return nil, fmt.Errorf("creating user: %w", err)
	}

	if claimRepo, ok := s.repo.(interface {
		ClaimUnlinkedAttendeesByEmail(
			ctx context.Context,
			userID uuid.UUID,
			email string,
		) (int64, error)
	}); ok {
		if _, err := claimRepo.ClaimUnlinkedAttendeesByEmail(ctx, user.ID, user.Email); err != nil {
			return nil, fmt.Errorf("claiming attendees by email: %w", err)
		}
	}

	return user, nil
}

// Login authenticates a user and returns a signed JWT.
func (s *AuthService) Login(
	ctx context.Context,
	email, password string,
) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("finding user: %w", err)
	}

	if user == nil {
		return "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := middleware.GenerateToken(
		*user, middleware.DefaultTokenTTL,
	)
	if err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}

	return token, nil
}

// GetUser returns a user by ID for session resolution.
func (s *AuthService) GetUser(
	ctx context.Context,
	id string,
) (*models.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetOrganizer is kept as a compatibility wrapper while handlers migrate.
func (s *AuthService) GetOrganizer(
	ctx context.Context,
	id string,
) (*models.User, error) {
	return s.GetUser(ctx, id)
}
