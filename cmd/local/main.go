// Entrypoint de DESENVOLVIMENTO LOCAL — expõe a mesma lógica de
// autenticação por CPF (internal/authflow) via um servidor net/http comum,
// sem precisar do runtime da AWS Lambda nem de nenhum deploy real. Só serve
// pra testar o fluxo completo (CPF → JWT → chamada autenticada no app)
// inteiramente na sua máquina, antes de ir pra AWS de verdade.
//
// Em produção quem responde é a Lambda de verdade (main.go, na raiz deste
// repositório), atrás do API Gateway — este arquivo nunca é implantado.
//
// Como usar (na raiz deste repositório):
//
//	export DB_HOST=localhost DB_PORT=5432 DB_USER=autoshop DB_PASSWORD=autoshop DB_NAME=autoshop DB_SSLMODE=disable
//	export JWT_SECRET=<o mesmo valor usado no .env do app-autoshop>
//	go run ./cmd/local
//
// (os valores de DB_* acima são os do Postgres do docker-compose do
// app-autoshop — suba-o primeiro com "docker compose up" lá, e cadastre
// pelo menos um cliente para ter um CPF válido pra testar.)
//
//	curl -X POST http://localhost:8081/auth -d '{"cpf":"<cpf de um cliente cadastrado>"}'
//
// O token retornado pode ser usado direto contra o app-autoshop:
//
//	curl http://localhost:8080/orders/customer/<customer_id> -H "Authorization: Bearer <token>"
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"lambda-auth/internal/authflow"
	"lambda-auth/internal/config"
)

type authRequest struct {
	CPF string `json:"cpf"`
}

func main() {
	db := config.NewPostgresPool()
	defer db.Close()

	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var body authRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "corpo da requisição inválido"})
			return
		}

		res := authflow.Authenticate(r.Context(), db, body.CPF)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(res.Status)
		_ = json.NewEncoder(w).Encode(res.Body)
	})

	log.Println("lambda-auth (dev local) ouvindo em :8081 — POST /auth {\"cpf\": \"...\"}")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
