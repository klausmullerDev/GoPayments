package worker

import (
	"context"
	"log"
	"time"

	"github.com/klausmullerDev/GoPayments/internal/service"
)

// BillingWorker executa periodicamente em segundo plano (Goroutine) a cobrança de assinaturas vencidas.
type BillingWorker struct {
	billingService *service.BillingService
	interval       time.Duration
}

// NewBillingWorker cria uma nova instância do worker de faturamento recorrente.
func NewBillingWorker(billingService *service.BillingService, interval time.Duration) *BillingWorker {
	return &BillingWorker{
		billingService: billingService,
		interval:       interval,
	}
}

// Start inicia o loop do worker em background até que o contexto seja cancelado (Graceful Shutdown).
func (w *BillingWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("⚙️ BillingWorker iniciado (intervalo de verificação: %v)", w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 BillingWorker encerrado com segurança (Graceful Shutdown)")
			return
		case <-ticker.C:
			summary, err := w.billingService.ProcessDueSubscriptions(ctx)
			if err != nil {
				log.Printf("❌ Erro no ciclo do BillingWorker: %v", err)
				continue
			}
			if summary.TotalDue > 0 {
				log.Printf(
					"💳 BillingWorker: %d vencidas | %d renovadas com sucesso | %d falhas",
					summary.TotalDue, summary.Processed, summary.Failed,
				)
			}
		}
	}
}

