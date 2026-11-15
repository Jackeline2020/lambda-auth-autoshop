// Package config lê a configuração da função a partir do ambiente.
package config

import "os"

// Env lê uma variável de ambiente, com valor padrão.
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
