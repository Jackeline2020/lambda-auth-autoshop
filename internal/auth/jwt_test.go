package auth_test

import (
	"lambda-auth/internal/auth"
	"testing"

	"github.com/stretchr/testify/assert"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, err := auth.GenerateToken("cust-1", "customer")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// confirma que o token gerado aqui é lido corretamente com o mesmo
	// formato de Claims usado pelo app-autoshop (pkg/auth/jwt.go) — as duas
	// structs precisam ficar em sincronia manualmente entre os repositórios.
	parsed, err := jwtlib.ParseWithClaims(token, &auth.Claims{}, func(tok *jwtlib.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	assert.NoError(t, err)

	claims, ok := parsed.Claims.(*auth.Claims)
	assert.True(t, ok)
	assert.Equal(t, "cust-1", claims.UserID)
	assert.Equal(t, "customer", claims.Role)
}

func TestGenerateToken_MissingSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	_, err := auth.GenerateToken("cust-1", "customer")

	assert.Error(t, err)
}
