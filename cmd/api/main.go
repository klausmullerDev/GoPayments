package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/klausmullerDev/GoPayments/internal/config"
	"github.com/klausmullerDev/GoPayments/internal/database"
	"github.com/klausmullerDev/GoPayments/internal/handler"
	"github.com/klausmullerDev/GoPayments/internal/middleware"
	"github.com/klausmullerDev/GoPayments/internal/repository"
	"github.com/klausmullerDev/GoPayments/internal/service"
)

func main() {
	// 1. Carregar variáveis de ambiente
	cfg := config.Load()

	// 2. Conectar ao banco de dados PostgreSQL via GORM e rodar Migrations
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("❌ Erro fatal ao inicializar banco de dados: %v", err)
	}
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("❌ Erro fatal nas migrations: %v", err)
	}

	// 3. Configurar modo do Gin e Router
	gin.SetMode(cfg.GinMode)
	router := gin.Default()
	_ = router.SetTrustedProxies(nil)

	// 4. Injeção de Dependências (Repositories -> Services -> Handlers)
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpirationHours)

	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(authService)

	// 5. Registrar rotas v1
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Check)

		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(authService), authHandler.Me)
		}
	}

	// 6. Iniciar servidor HTTP
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 GoPayments API rodando em http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ Erro ao iniciar servidor HTTP: %v", err)
	}
}
