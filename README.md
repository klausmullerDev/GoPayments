# 💳 GoPayments (GoPay) — SaaS Billing & Recurring Subscriptions API

API RESTful desenvolvida em **Go (Golang)** para gestão de planos, assinaturas recorrentes e faturamento automatizado (SaaS Billing).

Este projeto foi construído seguindo as melhores práticas de engenharia de software do mercado Go, utilizando **Arquitetura em Camadas (Clean/Layered Architecture)**, **Transações ACID com GORM**, **Autenticação JWT**, **Middlewares no Gin** e **Goroutines (Background Workers)** para processamento de cobranças recorrentes.

---

## 🚀 Tecnologias & Ferramentas

- **Linguagem:** Go 1.22+ (`github.com/klausmullerDev/GoPayments`)
- **Framework Web:** [Gin Gonic](https://github.com/gin-gonic/gin)
- **ORM & Banco de Dados:** [GORM](https://gorm.io/) + **PostgreSQL 16**
- **Autenticação & Segurança:** JWT (`golang-jwt/jwt/v5`) + Bcrypt (`golang.org/x/crypto/bcrypt`)
- **Concorrência:** Goroutines + `time.Ticker` / `context.Context` para faturamento em background
- **Containerização:** Docker & Docker Compose
- **Documentação:** Swagger / OpenAPI (`swaggo/swag`)
- **Testes:** `testing` standard library + mocks em memória

---

## 🏛️ Arquitetura do Projeto

O projeto adota o **Standard Go Project Layout** (`cmd/`, `internal/`, `docs/`) combinado com a **Arquitetura em Camadas com Injeção de Dependências via Interfaces**:

```text
GoPay/
├── cmd/
│   └── api/
│       └── main.go              # Entrypoint: inicializa configs, DB, Worker e servidor HTTP
├── internal/
│   ├── config/                  # Carregamento de variáveis de ambiente
│   ├── database/                # Conexão com PostgreSQL (GORM) e Auto-Migrations
│   ├── domain/                  # Entidades (Models), DTOs e Interfaces (Contratos)
│   ├── repository/              # Persistência de dados e Transações com GORM
│   ├── service/                 # Regras de negócio puras (UseCases)
│   ├── handler/                 # Controllers HTTP (Gin) e validação de entrada
│   ├── middleware/              # Middlewares de Autenticação JWT e Assinatura Ativa
│   └── worker/                  # Goroutine para cobrança recorrente de assinaturas
├── docs/                        # Documentação de Arquitetura (DER, Fluxos) e Swagger
├── docker-compose.yml           # Orquestração PostgreSQL + API
├── Dockerfile                   # Build multi-stage da aplicação Go
└── README.md
```

> 📖 Veja o detalhamento técnico completo, modelagem do banco de dados (DER) e diagramas de sequência em [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md).

---

## ⚙️ Funcionalidades Principais

1. **Autenticação & Perfis (`Users`):**
   - Cadastro de usuários (`CUSTOMER` e `ADMIN`) com senha criptografada via `bcrypt`.
   - Login com geração de token **JWT**.
2. **Catálogo de Planos (`Plans`):**
   - Criação, listagem e desativação de planos (Mensal, Anual) com preços armazenados em **centavos (`int64`)** para evitar erros de ponto flutuante (`float`).
3. **Ciclo de Vida de Assinaturas (`Subscriptions`):**
   - Adesão a um plano (impede múltiplas assinaturas ativas para o mesmo usuário).
   - Upgrade/Downgrade de plano.
   - Cancelamento de assinatura.
4. **Faturamento Recorrente & Transações (`Invoices` + `Billing Worker`):**
   - Geração de faturas com **Chave de Idempotência** (evita cobrança duplicada no mesmo ciclo).
   - Processamento atômico via `gorm.DB.Transaction`: cria a fatura, processa o pagamento simulado e atualiza a data da próxima cobrança (`next_billing_date`) ou marca a assinatura como inadimplente (`PAST_DUE`).
5. **Controle de Acesso por Assinatura (`Subscription Middleware`):**
   - Rota exclusiva para assinantes ativos que valida em tempo real se o usuário possui assinatura com status `ACTIVE`.

---

## 🛣️ Mapa de Endpoints (API v1)

| Método | Rota | Acesso | Descrição |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/health` | Público | Verifica status da API e conexão com o banco |
| `POST` | `/api/v1/auth/register` | Público | Cadastra um novo usuário |
| `POST` | `/api/v1/auth/login` | Público | Autentica usuário e retorna token JWT |
| `GET` | `/api/v1/plans` | Público | Lista todos os planos ativos |
| `POST` | `/api/v1/plans` | Admin (JWT) | Cria um novo plano de assinatura |
| `POST` | `/api/v1/subscriptions` | Cliente (JWT) | Assina um plano e gera a 1ª fatura |
| `GET` | `/api/v1/subscriptions/me` | Cliente (JWT) | Consulta a assinatura atual do usuário |
| `PATCH` | `/api/v1/subscriptions/cancel` | Cliente (JWT) | Cancela a renovação da assinatura |
| `GET` | `/api/v1/invoices/me` | Cliente (JWT) | Lista o histórico de faturas do usuário |
| `POST` | `/api/v1/billing/process` | Admin (JWT) | Aciona manualmente o ciclo de cobrança recorrente |
| `GET` | `/api/v1/content/premium` | Assinante Ativo | Rota de demonstração protegida pelo middleware de assinatura |

---

## 🌿 Fluxo de Desenvolvimento (Git Workflow)

Este projeto é desenvolvido utilizando **Feature Branches** e **Conventional Commits**:

- `main`: Código estável e documentado.
- `docs/architecture-planning`: Planejamento técnico, modelagem e documentação inicial.
- `chore/initial-setup`: Setup do módulo Go, Docker Compose, PostgreSQL, GORM e `/health`.
- `feat/users-auth`: Domínio de usuários, hash de senha e autenticação JWT.
- `feat/plans-crud`: Domínio e endpoints de planos de assinatura.
- `feat/subscriptions`: Adesão, consulta e cancelamento de assinaturas + transações.
- `feat/billing-worker`: Motor de faturamento recorrente (Goroutine) e rota exclusiva de assinante.
- `test/unit-and-docs`: Testes unitários nas camadas de `service` e documentação Swagger.
