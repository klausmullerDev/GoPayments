package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler gerencia o endpoint de verificação de integridade da API e do banco.
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler cria uma nova instância de HealthHandler.
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check responde ao GET /api/v1/health verificando a conectividade com o PostgreSQL.
func (h *HealthHandler) Check(c *gin.Context) {
	dbStatus := "up"
	httpCode := http.StatusOK

	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "down"
		httpCode = http.StatusServiceUnavailable
	}

	c.JSON(httpCode, gin.H{
		"service":   "GoPayments API",
		"status":    dbStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

