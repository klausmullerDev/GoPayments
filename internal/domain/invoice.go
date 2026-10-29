package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InvoiceStatus define os estados possíveis de uma fatura.
type InvoiceStatus string

const (
	InvoiceStatusPending InvoiceStatus = "PENDING"
	InvoiceStatusPaid    InvoiceStatus = "PAID"
	InvoiceStatusFailed  InvoiceStatus = "FAILED"
)

// Invoice representa uma fatura gerada por uma assinatura (GORM).
// Possui IdempotencyKey com uniqueIndex para impedir cobranças duplicadas no mesmo ciclo.
type Invoice struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	SubscriptionID uuid.UUID     `gorm:"type:uuid;index;not null" json:"subscription_id"`
	AmountCents    int64         `gorm:"not null" json:"amount_cents"`
	Status         InvoiceStatus `gorm:"type:varchar(20);not null" json:"status"`
	IdempotencyKey string        `gorm:"type:varchar(150);uniqueIndex;not null" json:"idempotency_key"`
	PaidAt         *time.Time    `json:"paid_at,omitempty"`
	DueDate        time.Time     `gorm:"not null" json:"due_date"`
	CreatedAt      time.Time     `json:"created_at"`
}

// BeforeCreate gera automaticamente o UUID da fatura antes de salvar no banco.
func (i *Invoice) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// InvoiceRepository define o contrato de persistência para consulta de faturas.
type InvoiceRepository interface {
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]Invoice, error)
}

