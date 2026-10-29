package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/klausmullerDev/GoPayments/internal/service"
)

// RequireActiveSubscription bloqueia o acesso à rota caso o usuário autenticado
// não possua uma assinatura com status ACTIVE.
func RequireActiveSubscription(subService *service.SubscriptionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "usuário não autenticado",
			})
			return
		}

		active, err := subService.HasActiveSubscription(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "erro ao verificar status da assinatura",
			})
			return
		}

		if !active {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "acesso negado: é necessário possuir uma assinatura ativa (ACTIVE)",
			})
			return
		}

		c.Next()
	}
}
