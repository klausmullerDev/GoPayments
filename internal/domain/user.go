package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole define os níveis de acesso do sistema.
type UserRole string

const (
	RoleCustomer UserRole = "CUSTOMER"
	RoleAdmin    UserRole = "ADMIN"
)

// User representa a entidade de usuário no banco de dados (GORM).
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(120);not null" json:"name"`
	Email        string    `gorm:"type:varchar(180);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Role         UserRole  `gorm:"type:varchar(20);not null;default:'CUSTOMER'" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// BeforeCreate é um hook do GORM executado antes de inserir o registro no banco.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Role == "" {
		u.Role = RoleCustomer
	}
	return nil
}

// RegisterInput representa o payload de entrada para cadastro de usuário.
type RegisterInput struct {
	Name     string   `json:"name" binding:"required,min=2,max=120"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=6,max=72"`
	Role     UserRole `json:"role" binding:"omitempty,oneof=CUSTOMER ADMIN"`
}

// LoginInput representa o payload de entrada para autenticação.
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse representa a resposta enviada após login ou registro bem-sucedido.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UserRepository define o contrato (interface) de persistência para usuários.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}
