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
      # O cadastro de clientes é do OS Service: a função consulta a rota
      # interna dele e não acessa nenhum banco (por isso não roda na VPC)
      OS_SERVICE_URL = var.os_service_url
      JWT_SECRET     = var.jwt_secret
    }
  }
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
