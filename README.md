# lambda-auth

Function serverless de autenticação por CPF, exigida pela Fase 3. Não é
login com senha: o cliente informa o CPF, a function confirma que existe um
cadastro com esse CPF no banco (mesma tabela `customers` usada pelo `app`) e
devolve um JWT — o mesmo formato de token que o `app` já sabe validar
(`pkg/auth/jwt.go`).

## Por que este repositório existe separado do `app`

A Fase 3 pede uma function serverless com deploy e CI/CD independentes do
resto da aplicação. Hoje este código ainda mora dentro do monorepo
`autoshop` (em `lambda-auth/`) — ver `docs/repo-split-plan.md` no repositório
principal para o plano completo de separação em 4 repositórios. Quando o
split acontecer de verdade, esta pasta vira o conteúdo integral deste
repositório, sem precisar reescrever nada: já tem `go.mod` próprio, código
isolado (nenhum import de `internal/` do app) e workflow de CI/CD próprio.

## Fluxo

```
POST {auth_endpoint} 
Content-Type: application/json

{"cpf": "529.982.247-25"}
```

Respostas:

| Situação | Status | Body |
|---|---|---|
| CPF válido e cadastrado | 200 | `{"token": "<jwt>"}` |
| CPF com formato/dígito verificador inválido | 400 | `{"message": "CPF inválido"}` |
| Corpo da requisição malformado | 400 | `{"message": "corpo da requisição inválido"}` |
| CPF válido mas não cadastrado | 404 | `{"message": "cliente não encontrado"}` |
| Erro de banco/token | 500 | `{"message": "..."}` |

O token gerado tem `user_id` (o ID do cliente) e `role: "customer"`, e
expira em 24h — mesma struct `Claims` do `pkg/auth/jwt.go` do app.

## Por que Lambda + API Gateway (e não outro serviço)

- **Lambda**: o caso de uso é uma chamada isolada, sem estado, de baixíssimo
  volume comparado à API principal — não justifica manter um serviço
  sempre ligado (e sem custo de execução, é serverless: só paga por
  invocação).
- **API Gateway (HTTP API, não REST API)**: só precisamos de uma rota
  (`POST /auth`) com proxy direto pra Lambda — o modelo REST API do API
  Gateway traz recursos (API keys, request validation embutido, etc.) que
  não usamos aqui e custam mais caro por requisição.
- **`provided.al2023` (custom runtime) em vez de um runtime Go gerenciado**:
  a AWS descontinuou o runtime gerenciado `go1.x` — hoje o caminho oficial
  pra Go é compilar um binário chamado `bootstrap` e rodar em cima do
  runtime customizado `provided.al2023`. É por isso que o build (ver CI/CD
  abaixo) gera um binário `bootstrap`, não um `.zip` de código-fonte.

## Variáveis de ambiente (injetadas pelo Terraform em runtime)

| Variável | Origem |
|---|---|
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | Lidas do secret `autoshop/rds/credentials` no Secrets Manager (criado pelo repositório `infra-db` / hoje `infra/aws/rds.tf`) via `data "aws_secretsmanager_secret_version"` |
| `DB_SSLMODE` | Fixo em `require` — a Lambda sempre fala com o RDS real, nunca com um Postgres local |
| `JWT_SECRET` | Passado como variável Terraform (`-var="jwt_secret=..."`), vinda do GitHub Secret `JWT_SECRET` deste repositório. **Precisa ter o mesmo valor no repositório `app`**, senão o app não consegue validar os tokens que esta Lambda emite |

## Dependência cross-repo: rede

A Lambda roda dentro da VPC (obrigatório pra alcançar o RDS, que não é
público). Isso cria um security group novo (`autoshop-lambda-auth`) que
precisa estar na allowlist de quem pode falar com o banco na porta 5432 —
hoje essa allowlist (`infra/aws/rds.tf`, `aws_security_group.db`) só libera
o security group dos nodes do EKS.

**Passo manual necessário depois do primeiro apply**: pegue o output
`lambda_security_group_id` deste projeto e cole na variável
`additional_db_ingress_security_group_ids` do projeto `infra/aws` (ou do
futuro repositório `infra-db`), depois rode `terraform apply` lá de novo.
Documentado também em `docs/repo-split-plan.md`, na mesma seção das outras
dependências manuais entre repositórios (mesmo padrão já usado pra
`IRSA_ROLE_ARN`).

## Build e deploy

```bash
go mod tidy
go build ./...
go test ./...

# build do binário pro runtime provided.al2023
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go

cd terraform
terraform init
terraform plan -var="jwt_secret=<mesmo-valor-do-app>"
terraform apply -var="jwt_secret=<mesmo-valor-do-app>"
```

O CI/CD (`.github/workflows/ci-cd.yml`) automatiza tudo isso: `test` roda
sempre, `terraform-validate` roda sempre, e `deploy` só aplica de verdade
se a variável de repositório `AWS_DEPLOY_ENABLED` estiver `true` — mesmo
padrão de segurança já usado no repositório principal, pra não gerar custo
AWS sem querer em todo push.

## O que falta pra isso virar um repositório de verdade

- Criar o repositório `lambda-auth` no GitHub (fora do escopo do que dá pra
  automatizar por aqui)
- Configurar os secrets do repositório: `AWS_ROLE_ARN` (role de OIDC —
  ainda não existe uma dedicada a este repo; o padrão seria replicar o
  `github-oidc.tf` que hoje vive em `infra/aws`) e `JWT_SECRET`
- Depois do primeiro apply, colar `lambda_security_group_id` em `infra/aws`
  (ver seção acima)
