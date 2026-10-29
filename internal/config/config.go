package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config armazena todas as variáveis de ambiente e configurações da aplicação.
type Config struct {
	Port               string
	GinMode            string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	JWTSecret          string
	JWTExpirationHours int
}

// Load carrega as variáveis do arquivo .env (se existir) ou do ambiente do sistema.
func Load() *Config {
	// Ignora erro caso o arquivo .env não exista (ex: rodando em container Docker puro)
	_ = godotenv.Load()

	expHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		expHours = 24
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		GinMode:            getEnv("GIN_MODE", "debug"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "gopay_user"),
		DBPassword:         getEnv("DB_PASSWORD", "gopay_secret"),
		DBName:             getEnv("DB_NAME", "gopay_db"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		JWTSecret:          getEnv("JWT_SECRET", "super_secret_jwt_key_change_in_production"),
		JWTExpirationHours: expHours,
	}
}

// DSN retorna a string de conexão formatada para o PostgreSQL (GORM).
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

