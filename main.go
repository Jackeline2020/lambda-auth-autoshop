package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"lambda-auth/internal/auth"
	"lambda-auth/internal/config"
	"lambda-auth/internal/customer"
	"lambda-auth/internal/validator"
)

type authRequest struct {
	CPF string `json:"cpf"`
}

type authResponse struct {
	Token string `json:"token"`
}

type errorResponse struct {
	Message string `json:"message"`
}

// db é criado uma única vez, fora do handler: o Lambda reaproveita o mesmo
// container (e portanto a mesma pool de conexão) entre invocações "quentes"
// — abrir uma conexão nova a cada chamada seria caro e desnecessário.
var db = config.NewPostgresPool()

// handler implementa o fluxo de autenticação por CPF exigido pela Fase 3:
// recebe o CPF, confirma que existe um cliente cadastrado com ele, e emite
// um JWT. Não há senha — a identificação é só o CPF, igual a um totem de
// autoatendimento.
func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var body authRequest
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return respond(http.StatusBadRequest, errorResponse{Message: "corpo da requisição inválido"})
	}

	if !validator.IsValidCPF(body.CPF) {
		return respond(http.StatusBadRequest, errorResponse{Message: "CPF inválido"})
	}

	repo := customer.NewRepository(db)
	c, err := repo.FindByCPF(ctx, validator.OnlyDigits(body.CPF))
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return respond(http.StatusNotFound, errorResponse{Message: "cliente não encontrado"})
		}
		return respond(http.StatusInternalServerError, errorResponse{Message: "erro ao consultar cliente"})
	}

	token, err := auth.GenerateToken(c.ID, "customer")
	if err != nil {
		return respond(http.StatusInternalServerError, errorResponse{Message: "erro ao gerar token"})
	}

	return respond(http.StatusOK, authResponse{Token: token})
}

func respond(status int, payload any) (events.APIGatewayV2HTTPResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, err
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
