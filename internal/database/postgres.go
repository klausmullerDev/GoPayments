package database

import (
	"fmt"
	"log"
	"time"

	"github.com/klausmullerDev/GoPayments/internal/config"
	"github.com/klausmullerDev/GoPayments/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB inicializa a conexão com o PostgreSQL utilizando GORM
// e configura o pool de conexões.
func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
	gormLogLevel := logger.Info
	if cfg.GinMode == "release" {
		gormLogLevel = logger.Warn
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("falha ao obter instância sql.DB: %w", err)
	}

	// Configuração do Connection Pool (boa prática de produção)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("falha no ping ao PostgreSQL: %w", err)
	}

	log.Println("✅ Conexão com PostgreSQL estabelecida com sucesso via GORM")
	return db, nil
}

// RunMigrations executa o AutoMigrate do GORM para criar/atualizar as tabelas do domínio.
func RunMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(&domain.User{}, &domain.Plan{}); err != nil {
		return fmt.Errorf("falha ao executar migrations: %w", err)
	}
	log.Println("✅ Migrations executadas com sucesso")
	return nil
}


