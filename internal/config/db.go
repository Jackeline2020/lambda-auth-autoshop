package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool cria a pool de conexão a partir de variáveis de ambiente.
// No Lambda, essas variáveis são injetadas pelo Terraform (terraform/main.tf)
// a partir do secret do RDS que o repositório infra-db provisiona — a Lambda
// só lê o secret, não cria nem gerencia o banco.
func NewPostgresPool() *pgxpool.Pool {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		getEnv("DB_USER", "autoshop"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "autoshop"),
		getEnv("DB_SSLMODE", "require"),
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("erro ao criar pool de conexão com o Postgres: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("erro ao conectar ao Postgres: %v", err)
	}

	return pool
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
