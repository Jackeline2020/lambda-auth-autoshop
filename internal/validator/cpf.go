package validator

import (
	"regexp"
	"strings"
)

var nonDigits = regexp.MustCompile(`\D`)

// OnlyDigits remove pontuação (pontos, traço) de um CPF, deixando só os
// dígitos. Necessário porque o CPF pode chegar formatado ("529.982.247-25")
// ou não ("52998224725"), e a tabela customers guarda o valor como foi
// cadastrado originalmente.
func OnlyDigits(cpf string) string {
	return nonDigits.ReplaceAllString(cpf, "")
}

// IsValidCPF replica o algoritmo de dígito verificador (mod 11) usado no
// repositório principal (pkg/validator/document.go). Copiado aqui — não
// reaproveitado por import — porque este é um módulo Go separado
// (lambda-auth), sem acesso ao pacote interno do app.
func IsValidCPF(cpf string) bool {
	digits := OnlyDigits(cpf)

	if len(digits) != 11 {
		return false
	}

	// Rejeita CPFs com todos os dígitos iguais (ex: 111.111.111-11)
	if strings.Count(digits, string(digits[0])) == 11 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (10 - i)
	}
	remainder := (sum * 10) % 11
	if remainder == 10 || remainder == 11 {
		remainder = 0
	}
	if remainder != int(digits[9]-'0') {
		return false
	}

	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(digits[i]-'0') * (11 - i)
	}
	remainder = (sum * 10) % 11
	if remainder == 10 || remainder == 11 {
		remainder = 0
	}

	return remainder == int(digits[10]-'0')
}
