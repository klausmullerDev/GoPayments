package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/klausmullerDev/GoPayments/internal/domain"
	"github.com/klausmullerDev/GoPayments/internal/middleware"
	"github.com/klausmullerDev/GoPayments/internal/service"
)

// AuthHandler gerencia as rotas HTTP de autenticação e perfil de usuário.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler cria uma nova instância de AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register trata o endpoint POST /api/v1/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var input domain.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "dados de entrada inválidos",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno ao registrar usuário"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login trata o endpoint POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var input domain.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "dados de entrada inválidos",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno ao realizar login"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Me trata o endpoint protegido GET /api/v1/auth/me retornando o perfil do usuário autenticado.
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	user, err := h.authService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar perfil do usuário"})
		return
	}

	c.JSON(http.StatusOK, user)
}

