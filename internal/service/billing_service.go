package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/klausmullerDev/GoPayments/internal/domain"
)

// BillingSummary representa o resultado consolidado de um ciclo de processamento de cobranças.
type BillingSummary struct {
	TotalDue    int       `json:"total_due"`
	Processed   int       `json:"processed"`
	Failed      int       `json:"failed"`
	ProcessedAt time.Time `json:"processed_at"`
}

// BillingService processa a cobrança recorrente das assinaturas vencidas.
type BillingService struct {
	subRepo domain.SubscriptionRepository
}

// NewBillingService cria uma nova instância de BillingService.
func NewBillingService(subRepo domain.SubscriptionRepository) *BillingService {
	return &BillingService{subRepo: subRepo}
}

// ProcessDueSubscriptions busca todas as assinaturas ativas cuja data de cobrança já venceu
// e executa a renovação atômica (nova fatura + avanço do next_billing_date).
func (s *BillingService) ProcessDueSubscriptions(ctx context.Context) (*BillingSummary, error) {
	now := time.Now().UTC()

	dueSubs, err := s.subRepo.FindDueSubscriptions(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar assinaturas vencidas: %w", err)
	}

	summary := &BillingSummary{
		TotalDue:    len(dueSubs),
		ProcessedAt: now,
	}

	for _, sub := range dueSubs {
		if sub.Plan == nil {
			summary.Failed++
			continue
		}

		// Chave determinística baseada no ciclo que estava agendado para vencer
		cycleDate := sub.NextBillingDate.UTC().Format("2006-01-02")
		idempotencyKey := fmt.Sprintf("sub_%s_cycle_%s", sub.ID.String(), cycleDate)

		invoice := &domain.Invoice{
			ID:             uuid.New(),
			SubscriptionID: sub.ID,
			AmountCents:    sub.Plan.PriceCents,
			Status:         domain.InvoiceStatusPaid,
			IdempotencyKey: idempotencyKey,
			PaidAt:         &now,
			DueDate:        sub.NextBillingDate,
		}

		nextBilling := CalculateNextBillingDate(now, sub.Plan.Interval)

		if err := s.subRepo.ProcessRenewalTransaction(ctx, sub.ID, invoice, now, nextBilling); err != nil {
			log.Printf("⚠️ Falha ao renovar assinatura %s: %v", sub.ID, err)
			summary.Failed++
			continue
		}

		summary.Processed++
	}

	return summary, nil
}

