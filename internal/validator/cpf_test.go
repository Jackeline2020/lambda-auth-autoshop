package validator

import "testing"

func TestIsValidCPF(t *testing.T) {
	tests := []struct {
		name string
		cpf  string
		want bool
	}{
		{"válido formatado", "529.982.247-25", true},
		{"válido sem formatação", "52998224725", true},
		{"válido formatado 2", "987.654.321-00", true},
		{"válido formatado 3", "135.792.468-28", true},
		{"todos dígitos iguais", "111.111.111-11", false},
		{"tamanho errado", "123.456.789-0", false},
		{"dígito verificador errado", "529.982.247-26", false},
		{"vazio", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidCPF(tt.cpf); got != tt.want {
				t.Errorf("IsValidCPF(%q) = %v, want %v", tt.cpf, got, tt.want)
			}
		})
	}
}

func TestOnlyDigits(t *testing.T) {
	if got := OnlyDigits("529.982.247-25"); got != "52998224725" {
		t.Errorf("OnlyDigits() = %q, want %q", got, "52998224725")
	}
}
