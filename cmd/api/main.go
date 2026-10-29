package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/klausmullerDev/GoPayments/internal/config"
	"github.com/klausmullerDev/GoPayments/internal/database"
	"github.com/klausmullerDev/GoPayments/internal/handler"
)

func main() {
	// 1. Carregar variáveis de ambiente
	cfg := config.Load()

	// 2. Conectar ao banco de dados PostgreSQL via GORM
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("❌ Erro fatal ao inicializar banco de dados: %v", err)
	}

	// 3. Configurar modo do Gin e Router
	gin.SetMode(cfg.GinMode)
	router := gin.Default()
	_ = router.SetTrustedProxies(nil)

	// 4. Inicializar Handlers e registrar rotas v1
	healthHandler := handler.NewHealthHandler(db)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Check)
	}

	// 5. Iniciar servidor HTTP
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 GoPayments API rodando em http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ Erro ao iniciar servidor HTTP: %v", err)
	}
}

