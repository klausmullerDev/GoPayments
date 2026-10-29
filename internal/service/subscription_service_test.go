package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/klausmullerDev/GoPayments/internal/domain"
)

type mockPlanRepo struct {
	plans map[uuid.UUID]*domain.Plan
}

func newMockPlanRepo() *mockPlanRepo {
	return &mockPlanRepo{plans: make(map[uuid.UUID]*domain.Plan)}
}

func (m *mockPlanRepo) Create(_ context.Context, plan *domain.Plan) error {
	m.plans[plan.ID] = plan
	return nil
}

func (m *mockPlanRepo) ListActive(_ context.Context) ([]domain.Plan, error) {
	var res []domain.Plan
	for _, p := range m.plans {
		if p.IsActive {
			res = append(res, *p)
		}
	}
	return res, nil
}

func (m *mockPlanRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Plan, error) {
	p, ok := m.plans[id]
	if !ok {
		return nil, domain.ErrPlanNotFound
	}
	return p, nil
}

func (m *mockPlanRepo) Deactivate(_ context.Context, id uuid.UUID) error {
	p, ok := m.plans[id]
	if !ok {
		return domain.ErrPlanNotFound
	}
	p.IsActive = false
	return nil
}

type mockSubscriptionRepo struct {
	subs     map[uuid.UUID]*domain.Subscription
	invoices []domain.Invoice
}

func newMockSubscriptionRepo() *mockSubscriptionRepo {
	return &mockSubscriptionRepo{
		subs: make(map[uuid.UUID]*domain.Subscription),
	}
}

func (m *mockSubscriptionRepo) CreateWithInitialInvoice(_ context.Context, sub *domain.Subscription, invoice *domain.Invoice) error {
	m.subs[sub.ID] = sub
	m.invoices = append(m.invoices, *invoice)
	return nil
}

func (m *mockSubscriptionRepo) FindActiveByUserID(_ context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	for _, s := range m.subs {
		if s.UserID == userID && s.Status == domain.SubStatusActive {
			return s, nil
		}
	}
	return nil, domain.ErrSubscriptionNotFound
}

func (m *mockSubscriptionRepo) FindLatestByUserID(_ context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	for _, s := range m.subs {
		if s.UserID == userID {
			return s, nil
		}
	}
	return nil, domain.ErrSubscriptionNotFound
}

func (m *mockSubscriptionRepo) Cancel(_ context.Context, subID uuid.UUID, canceledAt time.Time) error {
	s, ok := m.subs[subID]
	if !ok {
		return domain.ErrSubscriptionNotFound
	}
	s.Status = domain.SubStatusCanceled
	s.CanceledAt = &canceledAt
	return nil
}

func (m *mockSubscriptionRepo) FindDueSubscriptions(_ context.Context, now time.Time) ([]domain.Subscription, error) {
	var due []domain.Subscription
	for _, s := range m.subs {
		if s.Status == domain.SubStatusActive && !s.NextBillingDate.After(now) {
			due = append(due, *s)
		}
	}
	return due, nil
}

func (m *mockSubscriptionRepo) ProcessRenewalTransaction(
	_ context.Context,
	subID uuid.UUID,
	invoice *domain.Invoice,
	nextStart, nextBilling time.Time,
) error {
	s, ok := m.subs[subID]
	if !ok {
		return domain.ErrSubscriptionNotFound
	}
	m.invoices = append(m.invoices, *invoice)
	s.CurrentPeriodStart = nextStart
	s.NextBillingDate = nextBilling
	return nil
}

type mockInvoiceRepo struct{}

func (m *mockInvoiceRepo) ListByUserID(_ context.Context, _ uuid.UUID) ([]domain.Invoice, error) {
	return nil, nil
}

func TestSubscriptionAndBillingFlow(t *testing.T) {
	planRepo := newMockPlanRepo()
	subRepo := newMockSubscriptionRepo()
	invRepo := &mockInvoiceRepo{}

	subSvc := NewSubscriptionService(subRepo, planRepo, invRepo)
	billingSvc := NewBillingService(subRepo)
	ctx := context.Background()

	planID := uuid.New()
	_ = planRepo.Create(ctx, &domain.Plan{
		ID:         planID,
		Name:       "Pro Mensal",
		PriceCents: 4990,
		Interval:   domain.IntervalMonthly,
		IsActive:   true,
	})

	userID := uuid.New()

	// 1. Deve criar assinatura e fatura inicial com sucesso
	resp, err := subSvc.Subscribe(ctx, userID, domain.CreateSubscriptionInput{PlanID: planID})
	if err != nil {
		t.Fatalf("esperava sucesso ao assinar plano, obteve erro: %v", err)
	}
	if resp.Subscription.Status != domain.SubStatusActive {
		t.Errorf("esperava status ACTIVE, obteve %s", resp.Subscription.Status)
	}
	if resp.Invoice.AmountCents != 4990 || resp.Invoice.Status != domain.InvoiceStatusPaid {
		t.Errorf("fatura inicial inconsistente: %+v", resp.Invoice)
	}

	// 2. Deve bloquear segunda assinatura ativa para o mesmo usuário
	_, err = subSvc.Subscribe(ctx, userID, domain.CreateSubscriptionInput{PlanID: planID})
	if !errors.Is(err, domain.ErrActiveSubscriptionExists) {
		t.Fatalf("esperava ErrActiveSubscriptionExists, obteve: %v", err)
	}

	// 3. Simula vencimento da assinatura e processa cobrança recorrente
	subRepo.subs[resp.Subscription.ID].NextBillingDate = time.Now().UTC().Add(-24 * time.Hour)

	summary, err := billingSvc.ProcessDueSubscriptions(ctx)
	if err != nil {
		t.Fatalf("esperava sucesso no faturamento recorrente, obteve: %v", err)
	}
	if summary.TotalDue != 1 || summary.Processed != 1 {
		t.Fatalf("esperava 1 assinatura renovada, obteve: %+v", summary)
	}
	if len(subRepo.invoices) != 2 {
		t.Fatalf("esperava 2 faturas emitidas após renovação, obteve %d", len(subRepo.invoices))
	}

	// 4. Deve cancelar a assinatura com sucesso
	canceledSub, err := subSvc.CancelMySubscription(ctx, userID)
	if err != nil {
		t.Fatalf("esperava sucesso ao cancelar assinatura, obteve: %v", err)
	}
	if canceledSub.Status != domain.SubStatusCanceled || canceledSub.CanceledAt == nil {
		t.Errorf("assinatura deveria estar CANCELED com data preenchida: %+v", canceledSub)
	}
}
