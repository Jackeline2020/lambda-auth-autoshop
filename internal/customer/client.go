// Package customer consulta o cliente no OS Service, dono do cadastro. A
// função de autenticação não acessa o banco do OS (cada serviço com o
// próprio banco): usa a rota interna GET /internal/customers/by-cpf/{cpf},
// que exige um token de serviço.
package customer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"lambda-auth/internal/auth"
)

var ErrNotFound = errors.New("cliente não encontrado")

const StatusActive = "ativo"

type Customer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Client chama o OS Service.
type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// FindByCPF busca o cliente pelo CPF (só dígitos).
func (c *Client) FindByCPF(ctx context.Context, cpf string) (Customer, error) {
	token, err := auth.GenerateToken("lambda-auth", "service")
	if err != nil {
		return Customer{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/internal/customers/by-cpf/"+url.PathEscape(cpf), nil)
	if err != nil {
		return Customer{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return Customer{}, fmt.Errorf("OS Service indisponível: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var cust Customer
		if err := json.NewDecoder(resp.Body).Decode(&cust); err != nil {
			return Customer{}, fmt.Errorf("resposta inválida do OS Service: %w", err)
		}
		return cust, nil
	case http.StatusNotFound:
		return Customer{}, ErrNotFound
	default:
		return Customer{}, fmt.Errorf("OS Service respondeu %d", resp.StatusCode)
	}
}
