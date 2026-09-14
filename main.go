package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"lambda-auth/internal/authflow"
	"lambda-auth/internal/config"
)

type authRequest struct {
	CPF string `json:"cpf"`
}

type errorResponse struct {
	Message string `json:"message"`
}

// db é criado uma única vez, fora do handler: o Lambda reaproveita o mesmo
// container (e portanto a mesma pool de conexão) entre invocações "quentes"
// — abrir uma conexão nova a cada chamada seria caro e desnecessário.
var db = config.NewPostgresPool()

// handler é o entrypoint de produção (API Gateway → Lambda). A regra de
// negócio em si mora em internal/authflow — ver cmd/local/main.go pro
// entrypoint equivalente usado em desenvolvimento local.
func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var body authRequest
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return respond(http.StatusBadRequest, errorResponse{Message: "corpo da requisição inválido"})
	}

	res := authflow.Authenticate(ctx, db, body.CPF)
	return respond(res.Status, res.Body)
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
