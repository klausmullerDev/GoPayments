package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/klausmullerDev/GoPayments/internal/domain"
	"github.com/klausmullerDev/GoPayments/internal/service"
)

// PlanHandler gerencia os endpoints HTTP de planos de assinatura.
type PlanHandler struct {
	planService *service.PlanService
}

// NewPlanHandler cria uma nova instância de PlanHandler.
func NewPlanHandler(planService *service.PlanService) *PlanHandler {
	return &PlanHandler{planService: planService}
}

// List responde ao GET /api/v1/plans listando todos os planos ativos.
func (h *PlanHandler) List(c *gin.Context) {
	plans, err := h.planService.ListActivePlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar planos"})
		return
	}
	c.JSON(http.StatusOK, plans)
}

// GetByID responde ao GET /api/v1/plans/:id buscando um plano pelo UUID.
func (h *PlanHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do plano inválido (formato UUID esperado)"})
		return
	}

	plan, err := h.planService.GetPlanByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar plano"})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// Create responde ao POST /api/v1/plans (restrito a usuários ADMIN).
func (h *PlanHandler) Create(c *gin.Context) {
	var input domain.CreatePlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "dados de entrada inválidos",
			"details": err.Error(),
		})
		return
	}

	plan, err := h.planService.CreatePlan(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao criar plano"})
		return
	}

	c.JSON(http.StatusCreated, plan)
}

// Deactivate responde ao PATCH /api/v1/plans/:id/deactivate (restrito a usuários ADMIN).
func (h *PlanHandler) Deactivate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do plano inválido (formato UUID esperado)"})
		return
	}

	if err := h.planService.DeactivatePlan(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrPlanNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao desativar plano"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "plano desativado com sucesso"})
}
