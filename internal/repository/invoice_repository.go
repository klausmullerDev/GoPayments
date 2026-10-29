package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/klausmullerDev/GoPayments/internal/domain"
	"gorm.io/gorm"
)

type invoiceRepository struct {
	db *gorm.DB
}

// NewInvoiceRepository cria uma nova instância do repositório de faturas (GORM).
func NewInvoiceRepository(db *gorm.DB) domain.InvoiceRepository {
	return &invoiceRepository{db: db}
}

// ListByUserID retorna todas as faturas geradas pelas assinaturas de um determinado usuário.
func (r *invoiceRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Invoice, error) {
	var invoices []domain.Invoice
	err := r.db.WithContext(ctx).
		Joins("JOIN subscriptions ON subscriptions.id = invoices.subscription_id").
		Where("subscriptions.user_id = ?", userID).
		Order("invoices.created_at DESC").
		Find(&invoices).Error
	if err != nil {
		return nil, err
	}
	return invoices, nil
}

