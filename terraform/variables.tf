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

variable "db_secret_name" {
  description = "Nome do secret no Secrets Manager com as credenciais do RDS, criado pelo repositório infra-db (ver infra/aws/rds.tf: aws_secretsmanager_secret.db)"
  type        = string
  default     = "autoshop/rds/credentials"
}
