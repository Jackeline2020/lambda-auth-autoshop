package authflow_test

import (
	"context"
	"errors"
	"testing"

	"lambda-auth/internal/authflow"
	"lambda-auth/internal/customer"

	"github.com/stretchr/testify/assert"
)

type fakeFinder struct {
	c   customer.Customer
	err error
	got string
}

func (f *fakeFinder) FindByCPF(_ context.Context, cpf string) (customer.Customer, error) {
	f.got = cpf
	return f.c, f.err
}

func TestAuthenticate(t *testing.T) {
	t.Setenv("JWT_SECRET", "segredo-de-teste")
	ctx := context.Background()

	active := &fakeFinder{c: customer.Customer{ID: "c1", Status: customer.StatusActive}}
	res := authflow.Authenticate(ctx, active, "529.982.247-25")
	assert.Equal(t, 200, res.Status)
	assert.Equal(t, "52998224725", active.got, "o CPF vai só com dígitos")

	cases := []struct {
		name   string
		finder *fakeFinder
		cpf    string
		status int
	}{
		{"CPF inválido", &fakeFinder{}, "123", 400},
		{"cliente não encontrado", &fakeFinder{err: customer.ErrNotFound}, "529.982.247-25", 404},
		{"OS Service fora", &fakeFinder{err: errors.New("timeout")}, "529.982.247-25", 500},
		{"cliente inativo", &fakeFinder{c: customer.Customer{ID: "c1", Status: "inativo"}}, "529.982.247-25", 403},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.status, authflow.Authenticate(ctx, tc.finder, tc.cpf).Status)
		})
	}

	t.Setenv("JWT_SECRET", "")
	assert.Equal(t, 500, authflow.Authenticate(ctx, active, "529.982.247-25").Status)
}
