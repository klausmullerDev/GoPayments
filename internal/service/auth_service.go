package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/klausmullerDev/GoPayments/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// TokenClaims define os dados armazenados dentro do payload do JWT.
type TokenClaims struct {
	UserID uuid.UUID       `json:"user_id"`
	Email  string          `json:"email"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// AuthService concentra as regras de negócio de cadastro, login e validação JWT.
type AuthService struct {
	userRepo           domain.UserRepository
	jwtSecret          []byte
	jwtExpirationHours int
}

// NewAuthService cria uma nova instância de AuthService com injeção de dependência.
func NewAuthService(userRepo domain.UserRepository, jwtSecret string, jwtExpirationHours int) *AuthService {
	return &AuthService{
		userRepo:           userRepo,
		jwtSecret:          []byte(jwtSecret),
		jwtExpirationHours: jwtExpirationHours,
	}
}

// Register cadastra um novo usuário com senha criptografada via bcrypt e retorna o token JWT.
func (s *AuthService) Register(ctx context.Context, input domain.RegisterInput) (*domain.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	// Verifica se já existe usuário com este e-mail
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	// Gera hash seguro da senha com bcrypt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar hash da senha: %w", err)
	}

	role := input.Role
	if role == "" {
		role = domain.RoleCustomer
	}

	user := &domain.User{
		ID:           uuid.New(),
		Name:         strings.TrimSpace(input.Name),
		Email:        email,
		PasswordHash: string(hashedBytes),
		Role:         role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// Login autentica um usuário existente validando sua senha e emite um token JWT.
func (s *AuthService) Login(ctx context.Context, input domain.LoginInput) (*domain.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

// GetProfile busca os dados do usuário autenticado pelo ID.
func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// ValidateToken analisa e valida a assinatura e expiração de um token JWT.
func (s *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}

func (s *AuthService) generateToken(user *domain.User) (string, error) {
	now := time.Now().UTC()
	claims := TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.jwtExpirationHours) * time.Hour)),
			Issuer:    "gopay-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

