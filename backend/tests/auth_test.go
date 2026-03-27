package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/eventiofoss/eventio/backend/internal/models"
	"github.com/eventiofoss/eventio/backend/internal/service"
)

// --- mock repository --------------------------------------------------

type mockOrganizerRepo struct {
	organizers map[string]*models.Organizer
	createErr  error
}

func newMockRepo() *mockOrganizerRepo {
	return &mockOrganizerRepo{
		organizers: make(map[string]*models.Organizer),
	}
}

func (m *mockOrganizerRepo) Create(
	_ context.Context,
	org *models.Organizer,
) error {
	if m.createErr != nil {
		return m.createErr
	}

	if _, exists := m.organizers[org.Email]; exists {
		return errors.New("duplicate key value")
	}

	m.organizers[org.Email] = org
	return nil
}

func (m *mockOrganizerRepo) FindByEmail(
	_ context.Context,
	email string,
) (*models.Organizer, error) {
	org, ok := m.organizers[email]
	if !ok {
		return nil, nil
	}

	return org, nil
}

func (m *mockOrganizerRepo) FindByID(
	_ context.Context,
	id string,
) (*models.Organizer, error) {
	for _, org := range m.organizers {
		if org.ID.String() == id {
			return org, nil
		}
	}

	return nil, nil
}

// --- tests ------------------------------------------------------------

func TestAuthService_Register_HappyPath(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewAuthServiceWithRepo(repo)

	org, err := svc.Register(
		context.Background(),
		"Alice", "Alice@Example.COM", "secureP@ss1",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if org.Email != "alice@example.com" {
		t.Errorf("email not normalised: got %s", org.Email)
	}

	if org.Role != models.OrganizerRoleOrganizer {
		t.Errorf("unexpected role: got %s", org.Role)
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewAuthServiceWithRepo(repo)

	_, err := svc.Register(
		context.Background(),
		"Alice", "alice@example.com", "secureP@ss1",
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = svc.Register(
		context.Background(),
		"Alice2", "alice@example.com", "secureP@ss2",
	)
	if !errors.Is(err, service.ErrDuplicateEmail) {
		t.Errorf(
			"expected ErrDuplicateEmail, got: %v", err,
		)
	}
}

func TestAuthService_Register_InvalidInput(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewAuthServiceWithRepo(repo)

	cases := []struct {
		name    string
		n, e, p string
	}{
		{"empty name", "", "a@b.com", "password1"},
		{"empty email", "Ali", "", "password1"},
		{"short password", "Ali", "a@b.com", "short"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Register(
				context.Background(), tc.n, tc.e, tc.p,
			)
			if !errors.Is(err, service.ErrInvalidInput) {
				t.Errorf(
					"expected ErrInvalidInput, got: %v",
					err,
				)
			}
		})
	}
}

func TestAuthService_Login_HappyPath(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewAuthServiceWithRepo(repo)

	_, err := svc.Register(
		context.Background(),
		"Bob", "bob@example.com", "secureP@ss1",
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	token, err := svc.Login(
		context.Background(),
		"bob@example.com", "secureP@ss1",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewAuthServiceWithRepo(repo)

	_, err := svc.Register(
		context.Background(),
		"Carol", "carol@example.com", "secureP@ss1",
	)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, err = svc.Login(
		context.Background(),
		"carol@example.com", "wrongpassword",
	)
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf(
			"expected ErrInvalidCredentials, got: %v",
			err,
		)
	}
}

func TestAuthService_Login_NonExistentUser(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewAuthServiceWithRepo(repo)

	_, err := svc.Login(
		context.Background(),
		"nobody@example.com", "password1",
	)
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf(
			"expected ErrInvalidCredentials, got: %v",
			err,
		)
	}
}

func TestAuthService_GetOrganizer_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := service.NewAuthServiceWithRepo(repo)

	_, err := svc.GetOrganizer(
		context.Background(),
		"00000000-0000-0000-0000-000000000000",
	)
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf(
			"expected ErrInvalidCredentials, got: %v",
			err,
		)
	}
}
