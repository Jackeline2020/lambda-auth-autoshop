# Credenciais do RDS — criadas pelo repositório infra-db (Secrets Manager).
# lambda-auth só lê o secret existente, não cria nem gerencia o banco.
data "aws_secretsmanager_secret" "db" {
  name = var.db_secret_name
}

data "aws_secretsmanager_secret_version" "db" {
  secret_id = data.aws_secretsmanager_secret.db.id
}

locals {
  db_credentials = jsondecode(data.aws_secretsmanager_secret_version.db.secret_string)
}

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# --- Empacotamento do binário ---
# O binário "bootstrap" (runtime provided.al2023, ver README.md) precisa
# existir antes do apply — o workflow de CI/CD compila com
# GOOS=linux GOARCH=amd64 go build -o bootstrap main.go antes deste passo.
data "archive_file" "lambda" {
  type        = "zip"
  source_file = "${path.module}/../bootstrap"
  output_path = "${path.module}/lambda-auth.zip"
}

# --- IAM da execução da Lambda ---
resource "aws_iam_role" "lambda_exec" {
  name = "lambda-auth-exec"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Necessário porque a Lambda roda dentro da VPC (para alcançar o RDS, que
# não é público) — sem essa policy gerenciada, o Lambda não consegue criar
# as interfaces de rede (ENIs) exigidas pelo vpc_config abaixo.
resource "aws_iam_role_policy_attachment" "lambda_vpc" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

# --- Rede: a Lambda precisa estar na mesma VPC do RDS ---
resource "aws_security_group" "lambda" {
  name        = "autoshop-lambda-auth"
  description = "Allows the auth Lambda to reach Postgres (RDS) on port 5432"
  vpc_id      = data.aws_vpc.default.id

  egress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Project = "autoshop"
  }
}

# --- Lambda ---
resource "aws_lambda_function" "auth" {
  function_name = "autoshop-lambda-auth"
  role          = aws_iam_role.lambda_exec.arn

  filename         = data.archive_file.lambda.output_path
  source_code_hash = data.archive_file.lambda.output_base64sha256

  handler = "bootstrap"
  runtime = "provided.al2023"
  timeout = 10

  environment {
    variables = {
      DB_HOST     = local.db_credentials.host
      DB_PORT     = tostring(local.db_credentials.port)
      DB_NAME     = local.db_credentials.dbname
      DB_USER     = local.db_credentials.username
      DB_PASSWORD = local.db_credentials.password
      DB_SSLMODE  = "require"
      JWT_SECRET  = var.jwt_secret
    }
  }

  vpc_config {
    subnet_ids         = data.aws_subnets.default.ids
    security_group_ids = [aws_security_group.lambda.id]
  }

  depends_on = [aws_iam_role_policy_attachment.lambda_vpc]
}

# --- API Gateway (HTTP API — mais simples e mais barato que REST API pra
# esse caso de uso de uma rota só) ---
resource "aws_apigatewayv2_api" "auth" {
  name          = "autoshop-auth-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "auth" {
  api_id                 = aws_apigatewayv2_api.auth.id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.auth.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "auth" {
  api_id    = aws_apigatewayv2_api.auth.id
  route_key = "POST /auth"
  target    = "integrations/${aws_apigatewayv2_integration.auth.id}"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.auth.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_lambda_permission" "apigw" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.auth.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.auth.execution_arn}/*/*"
}
