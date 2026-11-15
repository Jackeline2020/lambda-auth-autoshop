package customer_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"lambda-auth/internal/customer"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindByCPF(t *testing.T) {
	t.Setenv("JWT_SECRET", "segredo-de-teste")

	var gotPath, gotRole string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		token := r.Header.Get("Authorization")[len("Bearer "):]
		claims := jwt.MapClaims{}
		_, _ = jwt.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) { return []byte("segredo-de-teste"), nil })
		gotRole, _ = claims["role"].(string)

		switch r.URL.Path {
		case "/internal/customers/by-cpf/52998224725":
			_, _ = w.Write([]byte(`{"id":"c1","name":"Ana","status":"ativo"}`))
		case "/internal/customers/by-cpf/11144477735":
			w.WriteHeader(http.StatusNotFound)
		case "/internal/customers/by-cpf/39053344705":
			_, _ = w.Write([]byte(`{`))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()
	client := customer.NewClient(srv.URL + "/")

	c, err := client.FindByCPF(context.Background(), "52998224725")
	require.NoError(t, err)
	assert.Equal(t, customer.Customer{ID: "c1", Name: "Ana", Status: "ativo"}, c)
	assert.Equal(t, "/internal/customers/by-cpf/52998224725", gotPath)
	assert.Equal(t, "service", gotRole, "a rota interna exige token de serviço")

	_, err = client.FindByCPF(context.Background(), "11144477735")
	assert.ErrorIs(t, err, customer.ErrNotFound)

	_, err = client.FindByCPF(context.Background(), "39053344705")
	assert.ErrorContains(t, err, "resposta inválida")

	_, err = client.FindByCPF(context.Background(), "00000000000")
	assert.ErrorContains(t, err, "500")
}

func TestFindByCPF_Errors(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	_, err := customer.NewClient("http://localhost:1").FindByCPF(context.Background(), "1")
	assert.ErrorContains(t, err, "JWT_SECRET")

	t.Setenv("JWT_SECRET", "x")
	_, err = customer.NewClient("http://127.0.0.1:1").FindByCPF(context.Background(), "1")
	assert.ErrorContains(t, err, "indisponível")

	_, err = customer.NewClient("http://[::1]:namedport").FindByCPF(context.Background(), "1")
	assert.Error(t, err)
}
