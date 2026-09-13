package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims replica a struct usada no repositório principal (pkg/auth/jwt.go)
// — o app valida os tokens com esse mesmo formato. O JWT_SECRET precisa ser
// o mesmo valor nos dois repositórios: colado manualmente como GitHub
// Secret em cada um, seguindo o mesmo padrão já usado para IRSA_ROLE_ARN.
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken é só a metade "emitir" do par usado no app (pkg/auth/jwt.go
// tem GenerateToken + ValidateToken). A Lambda só emite — quem valida é o
// próprio app, ao receber requisições com o token.
func GenerateToken(userID, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET não configurado")
	}

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
