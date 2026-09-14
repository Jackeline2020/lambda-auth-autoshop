package customer

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("cliente não encontrado")

type Customer struct {
	ID     string
	Name   string
	Status string
}

const StatusActive = "ativo"

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// FindByCPF busca o cliente pelo CPF já normalizado (só dígitos, ver
// validator.OnlyDigits). O regexp_replace do lado do banco existe porque a
// coluna cpf guarda o valor formatado como foi cadastrado (ex:
// "529.982.247-25") — comparar só dígitos evita falso-negativo por causa de
// pontuação.
func (r *Repository) FindByCPF(ctx context.Context, digitsOnlyCPF string) (Customer, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, status
		FROM customers
		WHERE regexp_replace(cpf, '\D', '', 'g') = $1
	`, digitsOnlyCPF)

	var c Customer
	if err := row.Scan(&c.ID, &c.Name, &c.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Customer{}, ErrNotFound
		}
		return Customer{}, err
	}
	return c, nil
}
