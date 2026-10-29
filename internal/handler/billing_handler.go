package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/klausmullerDev/GoPayments/internal/service"
)

// BillingHandler gerencia o disparo manual do faturamento recorrente e rotas exclusivas de assinantes.
type BillingHandler struct {
	billingService *service.BillingService
}

// NewBillingHandler cria uma nova instância de BillingHandler.
func NewBillingHandler(billingService *service.BillingService) *BillingHandler {
	return &BillingHandler{billingService: billingService}
}

// ProcessDue trata o endpoint administrativo POST /api/v1/billing/process.
func (h *BillingHandler) ProcessDue(c *gin.Context) {
	summary, err := h.billingService.ProcessDueSubscriptions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// PremiumContent trata o endpoint GET /api/v1/content/premium exclusivo para assinantes ativos.
func (h *BillingHandler) PremiumContent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"access":  "granted",
		"message": "🎉 Bem-vindo à área exclusiva de assinantes ativos do GoPayments!",
	})
}
