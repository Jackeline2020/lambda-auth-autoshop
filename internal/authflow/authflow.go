// Package authflow contém a regra de negócio do fluxo de autenticação por
// CPF exigido pela Fase 3: validar o CPF, confirmar a existência e o status
// do cliente cadastrado com ele, e emitir um JWT — sem senha, só o CPF,
// igual a um totem de autoatendimento.
//
// Fica isolado do entrypoint de propósito: é chamado tanto pelo handler
// Lambda de verdade (main.go, na raiz deste repositório — usado em
// produção via API Gateway) quanto pelo entrypoint local de desenvolvimento
// (cmd/local/main.go — um servidor net/http comum, só pra testar o fluxo
// inteiro sem precisar de um deploy real na AWS). As duas entradas chamam
// exatamente a mesma função, então nunca existem duas implementações da
// mesma regra que possam divergir.
package authflow

import (
	"context"
	"errors"

	"lambda-auth/internal/auth"
	"lambda-auth/internal/customer"
	"lambda-auth/internal/validator"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Response é neutra em relação ao transporte (HTTP comum ou API Gateway) —
// cada entrypoint converte isso pro formato que precisa.
type Response struct {
	Status int
	Body   any
}

type errorBody struct {
	Message string `json:"message"`
}

type tokenBody struct {
	Token string `json:"token"`
}

func Authenticate(ctx context.Context, db *pgxpool.Pool, cpf string) Response {
	if !validator.IsValidCPF(cpf) {
		return Response{Status: 400, Body: errorBody{Message: "CPF inválido"}}
	}

	repo := customer.NewRepository(db)
	c, err := repo.FindByCPF(ctx, validator.OnlyDigits(cpf))
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return Response{Status: 404, Body: errorBody{Message: "cliente não encontrado"}}
		}
		return Response{Status: 500, Body: errorBody{Message: "erro ao consultar cliente"}}
	}

	// Requisito da Fase 3: "consultar a existência e o status do cliente" —
	// existir não basta, um cliente inativo (ver internal/domain/customer.go
	// no app-autoshop) não recebe token mesmo com CPF válido e cadastrado.
	if c.Status != customer.StatusActive {
		return Response{Status: 403, Body: errorBody{Message: "cliente inativo — procure a oficina"}}
	}

	token, err := auth.GenerateToken(c.ID, "customer")
	if err != nil {
		return Response{Status: 500, Body: errorBody{Message: "erro ao gerar token"}}
	}

	return Response{Status: 200, Body: tokenBody{Token: token}}
}
