package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/klausmullerDev/GoPayments/internal/domain"
	"github.com/klausmullerDev/GoPayments/internal/service"
)

const (
	ContextUserIDKey   = "user_id"
	ContextUserRoleKey = "user_role"
)

// AuthMiddleware intercepta requisições HTTP exigindo um token Bearer JWT válido.
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "cabeçalho Authorization ausente",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "formato do token inválido (use: Bearer <token>)",
			})
			return
		}

		claims, err := authService.ValidateToken(strings.TrimSpace(parts[1]))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token inválido ou expirado",
			})
			return
		}

		// Armazena os dados do usuário autenticado no contexto do Gin
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserRoleKey, claims.Role)
		c.Next()
	}
}

// RequireRole verifica se o perfil (role) do usuário autenticado está entre os permitidos.
func RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get(ContextUserRoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "usuário não autenticado",
			})
			return
		}

		userRole, ok := val.(domain.UserRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "permissão inválida",
			})
			return
		}

		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "acesso negado: permissão insuficiente",
		})
	}
}

// GetUserID extrai com segurança o UUID do usuário autenticado a partir do contexto do Gin.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

