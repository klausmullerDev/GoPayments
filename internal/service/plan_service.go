package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/klausmullerDev/GoPayments/internal/domain"
)

// PlanService gerencia as regras de negócio do catálogo de planos.
type PlanService struct {
	planRepo domain.PlanRepository
}

// NewPlanService cria uma nova instância de PlanService.
func NewPlanService(planRepo domain.PlanRepository) *PlanService {
	return &PlanService{planRepo: planRepo}
}

// CreatePlan cadastra um novo plano de assinatura ativo.
func (s *PlanService) CreatePlan(ctx context.Context, input domain.CreatePlanInput) (*domain.Plan, error) {
	plan := &domain.Plan{
		ID:          uuid.New(),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		PriceCents:  input.PriceCents,
		Interval:    input.Interval,
		IsActive:    true,
	}

	if err := s.planRepo.Create(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// ListActivePlans retorna todos os planos ativos ordenados pelo menor preço.
func (s *PlanService) ListActivePlans(ctx context.Context) ([]domain.Plan, error) {
	return s.planRepo.ListActive(ctx)
}

// GetPlanByID busca um plano específico pelo seu UUID.
func (s *PlanService) GetPlanByID(ctx context.Context, id uuid.UUID) (*domain.Plan, error) {
	return s.planRepo.FindByID(ctx, id)
}

// DeactivatePlan desativa um plano para que novas assinaturas não possam escolhê-lo.
func (s *PlanService) DeactivatePlan(ctx context.Context, id uuid.UUID) error {
	return s.planRepo.Deactivate(ctx, id)
}

