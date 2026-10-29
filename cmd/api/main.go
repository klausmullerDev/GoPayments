package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klausmullerDev/GoPayments/internal/config"
	"github.com/klausmullerDev/GoPayments/internal/database"
	"github.com/klausmullerDev/GoPayments/internal/domain"
	"github.com/klausmullerDev/GoPayments/internal/handler"
	"github.com/klausmullerDev/GoPayments/internal/middleware"
	"github.com/klausmullerDev/GoPayments/internal/repository"
	"github.com/klausmullerDev/GoPayments/internal/service"
	"github.com/klausmullerDev/GoPayments/internal/worker"
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
	planRepo := repository.NewPlanRepository(db)
	subRepo := repository.NewSubscriptionRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
	planService := service.NewPlanService(planRepo)
	subService := service.NewSubscriptionService(subRepo, planRepo, invoiceRepo)
	billingService := service.NewBillingService(subRepo)

	healthHandler := handler.NewHealthHandler(db)
	authHandler := handler.NewAuthHandler(authService)
	planHandler := handler.NewPlanHandler(planService)
	subHandler := handler.NewSubscriptionHandler(subService)
	billingHandler := handler.NewBillingHandler(billingService)

	authMiddleware := middleware.AuthMiddleware(authService)

	// 5. Registrar rotas v1
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Check)

		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authMiddleware, authHandler.Me)
		}

		plans := v1.Group("/plans")
		{
			plans.GET("", planHandler.List)
			plans.GET("/:id", planHandler.GetByID)

			adminPlans := plans.Group("")
			adminPlans.Use(authMiddleware, middleware.RequireRole(domain.RoleAdmin))
			{
				adminPlans.POST("", planHandler.Create)
				adminPlans.PATCH("/:id/deactivate", planHandler.Deactivate)
			}
		}

		subscriptions := v1.Group("/subscriptions")
		subscriptions.Use(authMiddleware)
		{
			subscriptions.POST("", subHandler.Subscribe)
			subscriptions.GET("/me", subHandler.GetMe)
			subscriptions.PATCH("/cancel", subHandler.Cancel)
		}

		invoices := v1.Group("/invoices")
		invoices.Use(authMiddleware)
		{
			invoices.GET("/me", subHandler.ListMyInvoices)
		}

		// Motor de cobrança recorrente (acionamento manual para ADMIN)
		billing := v1.Group("/billing")
		billing.Use(authMiddleware, middleware.RequireRole(domain.RoleAdmin))
		{
			billing.POST("/process", billingHandler.ProcessDue)
		}

		// Rota protegida por assinatura ativa (JWT + RequireActiveSubscription)
		content := v1.Group("/content")
		content.Use(authMiddleware, middleware.RequireActiveSubscription(subService))
		{
			content.GET("/premium", billingHandler.PremiumContent)
		}
	}

	// 6. Iniciar Worker de cobrança recorrente em Goroutine com Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	billingWorker := worker.NewBillingWorker(billingService, 1*time.Minute)
	go billingWorker.Start(ctx)

	// 7. Iniciar servidor HTTP com suporte a Graceful Shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	go func() {
		log.Printf("🚀 GoPayments API rodando em http://localhost%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Erro ao iniciar servidor HTTP: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("🛑 Sinal de encerramento recebido, desligando servidor HTTP...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️ Erro no desligamento forçado do servidor: %v", err)
	}
	log.Println("✅ Servidor encerrado com sucesso")
}
