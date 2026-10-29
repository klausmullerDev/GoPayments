package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/klausmullerDev/GoPayments/internal/domain"
)

// SubscriptionService orquestra as regras de negócio de assinaturas e faturas.
type SubscriptionService struct {
	subRepo     domain.SubscriptionRepository
	planRepo    domain.PlanRepository
	invoiceRepo domain.InvoiceRepository
}

// NewSubscriptionService cria uma nova instância de SubscriptionService.
func NewSubscriptionService(
	subRepo domain.SubscriptionRepository,
	planRepo domain.PlanRepository,
	invoiceRepo domain.InvoiceRepository,
) *SubscriptionService {
	return &SubscriptionService{
		subRepo:     subRepo,
		planRepo:    planRepo,
		invoiceRepo: invoiceRepo,
	}
}

// Subscribe realiza a adesão de um usuário a um plano ativo, gerando a primeira fatura em transação ACID.
func (s *SubscriptionService) Subscribe(
	ctx context.Context,
	userID uuid.UUID,
	input domain.CreateSubscriptionInput,
) (*domain.SubscriptionResponse, error) {
	// 1. Verifica se o usuário já possui uma assinatura ativa
	existingSub, err := s.subRepo.FindActiveByUserID(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrSubscriptionNotFound) {
		return nil, err
	}
	if existingSub != nil {
		return nil, domain.ErrActiveSubscriptionExists
	}

	// 2. Busca o plano e valida se está ativo
	plan, err := s.planRepo.FindByID(ctx, input.PlanID)
	if err != nil {
		return nil, err
	}
	if !plan.IsActive {
		return nil, domain.ErrPlanNotFound
	}

	// 3. Calcula o ciclo de cobrança
	now := time.Now().UTC()
	nextBilling := CalculateNextBillingDate(now, plan.Interval)
	subID := uuid.New()

	sub := &domain.Subscription{
		ID:                 subID,
		UserID:             userID,
		PlanID:             plan.ID,
		Plan:               plan,
		Status:             domain.SubStatusActive,
		CurrentPeriodStart: now,
		NextBillingDate:    nextBilling,
	}

	// 4. Monta a fatura inicial com chave de idempotência determinística por ciclo
	idempotencyKey := fmt.Sprintf("sub_%s_cycle_%s", subID.String(), now.Format("2006-01-02"))
	invoice := &domain.Invoice{
		ID:             uuid.New(),
		SubscriptionID: subID,
		AmountCents:    plan.PriceCents,
		Status:         domain.InvoiceStatusPaid,
		IdempotencyKey: idempotencyKey,
		PaidAt:         &now,
		DueDate:        now,
	}

	// 5. Persiste Assinatura + Fatura em transação atômica no GORM
	if err := s.subRepo.CreateWithInitialInvoice(ctx, sub, invoice); err != nil {
		return nil, err
	}

	return &domain.SubscriptionResponse{
		Subscription: *sub,
		Invoice:      *invoice,
	}, nil
}

// GetMySubscription retorna a assinatura mais recente do usuário autenticado.
func (s *SubscriptionService) GetMySubscription(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	return s.subRepo.FindLatestByUserID(ctx, userID)
}

// CancelMySubscription cancela a renovação da assinatura ativa do usuário.
func (s *SubscriptionService) CancelMySubscription(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	sub, err := s.subRepo.FindLatestByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if sub.Status == domain.SubStatusCanceled {
		return nil, domain.ErrSubscriptionAlreadyCancel
	}

	now := time.Now().UTC()
	if err := s.subRepo.Cancel(ctx, sub.ID, now); err != nil {
		return nil, err
	}

	sub.Status = domain.SubStatusCanceled
	sub.CanceledAt = &now
	return sub, nil
}

// ListMyInvoices retorna o histórico de faturas emitidas para o usuário.
func (s *SubscriptionService) ListMyInvoices(ctx context.Context, userID uuid.UUID) ([]domain.Invoice, error) {
	return s.invoiceRepo.ListByUserID(ctx, userID)
}

// HasActiveSubscription verifica se o usuário possui uma assinatura com status ACTIVE.
func (s *SubscriptionService) HasActiveSubscription(ctx context.Context, userID uuid.UUID) (bool, error) {
	sub, err := s.subRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return false, nil
		}
		return false, err
	}
	return sub != nil && sub.Status == domain.SubStatusActive, nil
}

// CalculateNextBillingDate calcula a próxima data de vencimento com base no intervalo do plano.
func CalculateNextBillingDate(from time.Time, interval domain.PlanInterval) time.Time {
	if interval == domain.IntervalYearly {
		return from.AddDate(1, 0, 0)
	}
	return from.AddDate(0, 1, 0)
}

