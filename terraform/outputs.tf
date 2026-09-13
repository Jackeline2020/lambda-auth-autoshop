output "auth_endpoint" {
  description = "URL pública do endpoint de autenticação: POST {valor}/auth com body {\"cpf\": \"...\"}"
  value       = "${aws_apigatewayv2_stage.default.invoke_url}/auth"
}

output "lambda_security_group_id" {
  description = "ID do security group da Lambda. Precisa ser colado em infra/aws (variável additional_db_ingress_security_group_ids) para o RDS aceitar conexões vindas dela — o mesmo tipo de dependência manual entre repositórios já documentado em docs/repo-split-plan.md."
  value       = aws_security_group.lambda.id
}
