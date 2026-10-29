package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/klausmullerDev/GoPayments/internal/domain"
	"gorm.io/gorm"
)

type subscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository cria uma nova instância do repositório de assinaturas (GORM).
func NewSubscriptionRepository(db *gorm.DB) domain.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

// CreateWithInitialInvoice executa uma transação ACID criando a assinatura e sua primeira fatura.
// Se qualquer uma das operações falhar, o GORM executa ROLLBACK automático.
func (r *subscriptionRepository) CreateWithInitialInvoice(ctx context.Context, sub *domain.Subscription, invoice *domain.Invoice) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(sub).Error; err != nil {
			return err
		}
		if err := tx.Create(invoice).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *subscriptionRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("user_id = ? AND status = ?", userID, domain.SubStatusActive).
		First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) FindLatestByUserID(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) Cancel(ctx context.Context, subID uuid.UUID, canceledAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Subscription{}).
		Where("id = ?", subID).
		Updates(map[string]interface{}{
			"status":      domain.SubStatusCanceled,
			"canceled_at": canceledAt,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrSubscriptionNotFound
	}
	return nil
}

func (r *subscriptionRepository) FindDueSubscriptions(ctx context.Context, now time.Time) ([]domain.Subscription, error) {
	var subs []domain.Subscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("status = ? AND next_billing_date <= ?", domain.SubStatusActive, now).
		Find(&subs).Error
	if err != nil {
		return nil, err
	}
	return subs, nil
}

// ProcessRenewalTransaction executa atomicamente a criação da fatura de renovação
// e o avanço da data de próxima cobrança (next_billing_date) da assinatura.
func (r *subscriptionRepository) ProcessRenewalTransaction(
	ctx context.Context,
	subID uuid.UUID,
	invoice *domain.Invoice,
	nextStart, nextBilling time.Time,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(invoice).Error; err != nil {
			// Caso a chave de idempotência já exista, ignoramos duplicidade com segurança
			if strings.Contains(err.Error(), "duplicate key") || errors.Is(err, gorm.ErrDuplicatedKey) {
				return nil
			}
			return err
		}

		if err := tx.Model(&domain.Subscription{}).
			Where("id = ?", subID).
			Updates(map[string]interface{}{
				"current_period_start": nextStart,
				"next_billing_date":    nextBilling,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}
