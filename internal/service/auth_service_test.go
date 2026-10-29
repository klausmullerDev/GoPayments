package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/klausmullerDev/GoPayments/internal/domain"
)

// mockUserRepository implementa domain.UserRepository em memória para testes unitários.
type mockUserRepository struct {
	usersByEmail map[string]*domain.User
	usersByID    map[uuid.UUID]*domain.User
}

func newMockUserRepo() *mockUserRepository {
	return &mockUserRepository{
		usersByEmail: make(map[string]*domain.User),
		usersByID:    make(map[uuid.UUID]*domain.User),
	}
}

func (m *mockUserRepository) Create(_ context.Context, user *domain.User) error {
	if _, exists := m.usersByEmail[user.Email]; exists {
		return domain.ErrEmailAlreadyExists
	}
	m.usersByEmail[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *mockUserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	u, exists := m.usersByEmail[email]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepository) FindByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, exists := m.usersByID[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := newMockUserRepo()
	authSvc := NewAuthService(repo, "test-secret-key", 24)
	ctx := context.Background()

	// 1. Deve registrar um novo usuário com sucesso
	regResp, err := authSvc.Register(ctx, domain.RegisterInput{
		Name:     "Klaus Tester",
		Email:    "klaus@test.com",
		Password: "strongpassword",
	})
	if err != nil {
		t.Fatalf("esperava sucesso no registro, obteve erro: %v", err)
	}
	if regResp.Token == "" {
		t.Fatal("esperava token JWT preenchido")
	}
	if regResp.User.Role != domain.RoleCustomer {
		t.Errorf("esperava role CUSTOMER, obteve %s", regResp.User.Role)
	}

	// 2. Deve impedir cadastro com e-mail duplicado
	_, err = authSvc.Register(ctx, domain.RegisterInput{
		Name:     "Outro Nome",
		Email:    "klaus@test.com",
		Password: "strongpassword",
	})
	if !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Fatalf("esperava ErrEmailAlreadyExists, obteve: %v", err)
	}

	// 3. Deve autenticar com credenciais válidas
	loginResp, err := authSvc.Login(ctx, domain.LoginInput{
		Email:    "klaus@test.com",
		Password: "strongpassword",
	})
	if err != nil {
		t.Fatalf("esperava login com sucesso, obteve erro: %v", err)
	}

	// 4. Deve validar o token JWT gerado
	claims, err := authSvc.ValidateToken(loginResp.Token)
	if err != nil {
		t.Fatalf("esperava token válido, obteve erro: %v", err)
	}
	if claims.UserID != regResp.User.ID {
		t.Errorf("esperava UserID %s, obteve %s", regResp.User.ID, claims.UserID)
	}

	// 5. Deve rejeitar senha incorreta
	_, err = authSvc.Login(ctx, domain.LoginInput{
		Email:    "klaus@test.com",
		Password: "wrongpassword",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("esperava ErrInvalidCredentials, obteve: %v", err)
	}
}
