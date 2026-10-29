package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlanInterval define a periodicidade de cobrança do plano.
type PlanInterval string

const (
	IntervalMonthly PlanInterval = "MONTHLY"
	IntervalYearly  PlanInterval = "YEARLY"
)

// Plan representa um plano de assinatura no catálogo do sistema (GORM).
// PriceCents é armazenado em centavos (int64) para evitar erros de precisão com ponto flutuante.
type Plan struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string       `gorm:"type:varchar(100);not null" json:"name"`
	Description string       `gorm:"type:varchar(255)" json:"description"`
	PriceCents  int64        `gorm:"not null" json:"price_cents"`
	Interval    PlanInterval `gorm:"type:varchar(20);not null" json:"interval"`
	IsActive    bool         `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// BeforeCreate gera automaticamente o UUID do plano antes de inserir no banco.
func (p *Plan) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// CreatePlanInput representa o payload para criação de um novo plano (exclusivo ADMIN).
type CreatePlanInput struct {
	Name        string       `json:"name" binding:"required,min=2,max=100"`
	Description string       `json:"description" binding:"max=255"`
	PriceCents  int64        `json:"price_cents" binding:"required,gt=0"`
	Interval    PlanInterval `json:"interval" binding:"required,oneof=MONTHLY YEARLY"`
}

// PlanRepository define o contrato de persistência para planos de assinatura.
type PlanRepository interface {
	Create(ctx context.Context, plan *Plan) error
	ListActive(ctx context.Context) ([]Plan, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Plan, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
}
