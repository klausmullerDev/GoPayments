package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionStatus define o estado atual de uma assinatura.
type SubscriptionStatus string

const (
	SubStatusActive   SubscriptionStatus = "ACTIVE"
	SubStatusPastDue  SubscriptionStatus = "PAST_DUE"
	SubStatusCanceled SubscriptionStatus = "CANCELED"
)

// Subscription representa o vínculo de assinatura entre um Usuário e um Plano (GORM).
type Subscription struct {
	ID                 uuid.UUID          `gorm:"type:uuid;primaryKey" json:"id"`
	UserID             uuid.UUID          `gorm:"type:uuid;index;not null" json:"user_id"`
	User               *User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PlanID             uuid.UUID          `gorm:"type:uuid;index;not null" json:"plan_id"`
	Plan               *Plan              `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
	Status             SubscriptionStatus `gorm:"type:varchar(20);index;not null" json:"status"`
	CurrentPeriodStart time.Time          `gorm:"not null" json:"current_period_start"`
	NextBillingDate    time.Time          `gorm:"index;not null" json:"next_billing_date"`
	CanceledAt         *time.Time         `json:"canceled_at,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// BeforeCreate gera automaticamente o UUID da assinatura antes de inserir no banco.
func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// CreateSubscriptionInput representa o payload para assinar um plano.
type CreateSubscriptionInput struct {
	PlanID uuid.UUID `json:"plan_id" binding:"required"`
}

// SubscriptionResponse retorna a assinatura criada acompanhada da primeira fatura emitida.
type SubscriptionResponse struct {
	Subscription Subscription `json:"subscription"`
	Invoice      Invoice      `json:"initial_invoice"`
}

// SubscriptionRepository define o contrato de persistência e transações para assinaturas.
type SubscriptionRepository interface {
	CreateWithInitialInvoice(ctx context.Context, sub *Subscription, invoice *Invoice) error
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*Subscription, error)
	FindLatestByUserID(ctx context.Context, userID uuid.UUID) (*Subscription, error)
	Cancel(ctx context.Context, subID uuid.UUID, canceledAt time.Time) error
	FindDueSubscriptions(ctx context.Context, now time.Time) ([]Subscription, error)
	ProcessRenewalTransaction(ctx context.Context, subID uuid.UUID, invoice *Invoice, nextStart, nextBilling time.Time) error
}
