variable "aws_region" {
  description = "Região AWS onde os recursos serão criados"
  type        = string
  default     = "us-east-1"
}

variable "jwt_secret" {
  description = "Segredo usado para assinar os JWTs. Precisa ser idêntico ao configurado no repositório app — colado manualmente como GitHub Secret JWT_SECRET nos dois repositórios."
  type        = string
  sensitive   = true
}

variable "os_service_url" {
  description = "Endereço do OS Service (app-autoshop), consultado na rota interna GET /internal/customers/by-cpf/{cpf}"
  type        = string
}
