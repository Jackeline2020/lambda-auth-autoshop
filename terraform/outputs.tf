output "auth_endpoint" {
  description = "URL pública do endpoint de autenticação: POST {valor}/auth com body {\"cpf\": \"...\"}"
  value       = "${aws_apigatewayv2_stage.default.invoke_url}/auth"
}
