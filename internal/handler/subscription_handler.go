package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/klausmullerDev/GoPayments/internal/domain"
	"github.com/klausmullerDev/GoPayments/internal/middleware"
	"github.com/klausmullerDev/GoPayments/internal/service"
)

// SubscriptionHandler gerencia as rotas HTTP de assinaturas e faturas do cliente.
type SubscriptionHandler struct {
	subService *service.SubscriptionService
}

// NewSubscriptionHandler cria uma nova instância de SubscriptionHandler.
func NewSubscriptionHandler(subService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subService: subService}
}

// Subscribe trata o endpoint POST /api/v1/subscriptions.
func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	var input domain.CreateSubscriptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "dados de entrada inválidos",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.subService.Subscribe(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrActiveSubscriptionExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrPlanNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao criar assinatura"})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetMe trata o endpoint GET /api/v1/subscriptions/me.
func (h *SubscriptionHandler) GetMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	sub, err := h.subService.GetMySubscription(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar assinatura"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// Cancel trata o endpoint PATCH /api/v1/subscriptions/cancel.
func (h *SubscriptionHandler) Cancel(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	sub, err := h.subService.CancelMySubscription(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSubscriptionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrSubscriptionAlreadyCancel):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao cancelar assinatura"})
		}
		return
	}

	c.JSON(http.StatusOK, sub)
}

// ListMyInvoices trata o endpoint GET /api/v1/invoices/me.
func (h *SubscriptionHandler) ListMyInvoices(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	invoices, err := h.subService.ListMyInvoices(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar faturas"})
		return
	}

	c.JSON(http.StatusOK, invoices)
}

