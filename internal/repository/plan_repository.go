package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/klausmullerDev/GoPayments/internal/domain"
	"gorm.io/gorm"
)

type planRepository struct {
	db *gorm.DB
}

// NewPlanRepository cria uma nova instância do repositório de planos (GORM).
func NewPlanRepository(db *gorm.DB) domain.PlanRepository {
	return &planRepository{db: db}
}

func (r *planRepository) Create(ctx context.Context, plan *domain.Plan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *planRepository) ListActive(ctx context.Context) ([]domain.Plan, error) {
	var plans []domain.Plan
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("price_cents ASC").
		Find(&plans).Error
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *planRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Plan, error) {
	var plan domain.Plan
	err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPlanNotFound
		}
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Plan{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrPlanNotFound
	}
	return nil
}

