# 💳 GoPayments (GoPay) — SaaS Billing & Recurring Subscriptions API

API RESTful desenvolvida em **Go (Golang)** para gestão de planos, assinaturas recorrentes e faturamento automatizado (SaaS Billing).

Este projeto foi construído seguindo as melhores práticas de engenharia de software do mercado Go, utilizando **Arquitetura em Camadas com Injeção de Dependências via Interfaces**, **Transações ACID com GORM**, **Autenticação JWT**, **Middlewares no Gin** e **Goroutines (Background Workers)** com **Graceful Shutdown** para processamento de cobranças recorrentes.

---

## 🚀 Tecnologias & Ferramentas

- **Linguagem:** Go (`github.com/klausmullerDev/GoPayments`)
- **Framework Web:** [Gin Gonic](https://github.com/gin-gonic/gin)
- **ORM & Banco de Dados:** [GORM](https://gorm.io/) + **PostgreSQL 16**
- **Autenticação & Segurança:** JWT (`golang-jwt/jwt/v5`) + Bcrypt (`golang.org/x/crypto/bcrypt`)
- **Concorrência & Resiliência:** Goroutines + `time.Ticker` + `signal.NotifyContext` (Graceful Shutdown)
- **Containerização:** Docker (Multi-stage Build) & Docker Compose
- **Testes Automatizados:** `testing` standard library + In-Memory Mocks

---

## 🏛️ Arquitetura do Projeto

O projeto adota o **Standard Go Project Layout** (`cmd/`, `internal/`, `docs/`) combinado com a **Arquitetura em Camadas com Injeção de Dependências via Interfaces**:

```text
GoPay/
├── cmd/
│   └── api/
│       └── main.go                        # Entrypoint: configs, DB, Worker, Gin e Graceful Shutdown
├── internal/
│   ├── config/                            # Carregamento de variáveis de ambiente (.env)
│   ├── database/                          # Conexão com PostgreSQL (GORM), Pool e Auto-Migrations
│   ├── domain/                            # Entidades (Models), DTOs, Erros e Interfaces (Contratos)
│   ├── repository/                        # Persistência de dados e Transações ACID com GORM
│   ├── service/                           # Regras de negócio puras (UseCases) + Testes Unitários
│   ├── handler/                           # Controllers HTTP (Gin) e validação de entrada
│   ├── middleware/                        # Middlewares de Autenticação JWT, RBAC e Assinatura Ativa
│   └── worker/                            # Goroutine para cobrança recorrente de assinaturas
├── docs/
│   └── ARCHITECTURE.md                    # DER, Diagramas de Sequência e Decisões Técnicas
├── docker-compose.yml                     # Orquestração PostgreSQL 16 + API Go
├── Dockerfile                             # Build multi-stage otimizado (Alpine)
└── README.md
```

> 📖 Veja o detalhamento técnico completo, modelagem do banco de dados (DER) e diagramas de sequência em [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md).

---

## ⚙️ Funcionalidades Principais

1. **Autenticação & Perfis (`Users`):**
   - Cadastro de usuários (`CUSTOMER` e `ADMIN`) com senha criptografada via `bcrypt`.
   - Login com geração de token **JWT (`HS256`)** e rota `/me` para consulta de perfil.
2. **Catálogo de Planos (`Plans`):**
   - Criação, listagem, busca por ID e desativação de planos (`MONTHLY` ou `YEARLY`) com preços armazenados em **centavos (`int64`)** para evitar erros de ponto flutuante (`float`).
3. **Ciclo de Vida de Assinaturas (`Subscriptions`):**
   - Adesão a um plano (impede múltiplas assinaturas ativas para o mesmo usuário com `409 Conflict`).
   - Criação atômica de Assinatura + 1ª Fatura Paga via **GORM Transaction (`db.Transaction`)**.
   - Cancelamento de assinatura.
4. **Faturamento Recorrente & Idempotência (`Invoices` + `BillingWorker`):**
   - Geração de faturas com **Chave de Idempotência** (`sub_<id>_cycle_<YYYY-MM-DD>` com índice `UNIQUE` no PostgreSQL), evitando cobrança duplicada.
   - Goroutine em background que renova assinaturas vencidas e avança o `next_billing_date`.
5. **Controle de Acesso por Assinatura (`RequireActiveSubscription`):**
   - Middleware customizado no Gin que valida em tempo real se o usuário possui assinatura com status `ACTIVE` antes de liberar conteúdos exclusivos.

---

## 🛣️ Mapa de Endpoints (API v1)

| Método | Rota | Acesso | Descrição |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/health` | Público | Verifica status da API e ping no PostgreSQL |
| `POST` | `/api/v1/auth/register` | Público | Cadastra um novo usuário (`CUSTOMER` ou `ADMIN`) |
| `POST` | `/api/v1/auth/login` | Público | Autentica usuário e retorna token JWT |
| `GET` | `/api/v1/auth/me` | Autenticado (JWT) | Retorna dados do usuário logado |
| `GET` | `/api/v1/plans` | Público | Lista todos os planos ativos ordenados por preço |
| `GET` | `/api/v1/plans/:id` | Público | Busca detalhes de um plano pelo UUID |
| `POST` | `/api/v1/plans` | Admin (JWT) | Cria um novo plano de assinatura |
| `PATCH` | `/api/v1/plans/:id/deactivate` | Admin (JWT) | Desativa um plano para novas adesões |
| `POST` | `/api/v1/subscriptions` | Cliente (JWT) | Assina um plano e gera a 1ª fatura em transação ACID |
| `GET` | `/api/v1/subscriptions/me` | Cliente (JWT) | Consulta a assinatura atual do usuário |
| `PATCH` | `/api/v1/subscriptions/cancel` | Cliente (JWT) | Cancela a renovação da assinatura |
| `GET` | `/api/v1/invoices/me` | Cliente (JWT) | Lista o histórico de faturas do usuário |
| `POST` | `/api/v1/billing/process` | Admin (JWT) | Aciona sob demanda o ciclo de cobrança recorrente |
| `GET` | `/api/v1/content/premium` | Assinante Ativo | Rota protegida pelo middleware de assinatura ativa |

---

## ▶️ Como Executar o Projeto

### Opção 1: Rodando Tudo com Docker Compose (Recomendado)
Suba o banco PostgreSQL e a API Go com um único comando:
```bash
docker compose up --build -d
```
A API estará disponível em `http://localhost:8080/api/v1/health`.

### Opção 2: Rodando o Banco no Docker e a API Localmente
```bash
# 1. Copie o arquivo de variáveis de ambiente
cp .env.example .env

# 2. Suba apenas o PostgreSQL
docker compose up -d postgres

# 3. Execute a API
go run ./cmd/api
```

### 🧪 Executando os Testes Unitários
Os testes unitários da camada `service` utilizam *mocks* em memória das interfaces do domínio (não dependem de banco de dados):
```bash
go test -v ./...
```

---

## 💻 Testando o Fluxo Completo via `curl`

### 1. Criar um usuário ADMIN e cadastrar um Plano
```bash
# Registrar Admin
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Admin GoPay","email":"admin@gopay.dev","password":"adminpassword","role":"ADMIN"}'

# Criar Plano Pro (R$ 59,90 = 5990 centavos) usando o token do Admin
curl -X POST http://localhost:8080/api/v1/plans \
  -H "Authorization: Bearer <TOKEN_ADMIN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"GoPay Pro","description":"Acesso completo mensal","price_cents":5990,"interval":"MONTHLY"}'
```

### 2. Criar um Cliente, Assinar um Plano e Acessar Conteúdo Exclusivo
```bash
# Registrar Cliente
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Klaus Muller","email":"klaus@gopay.dev","password":"password123"}'

# Assinar o plano (gera Assinatura + Fatura inicial em transação ACID)
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Authorization: Bearer <TOKEN_CLIENTE>" \
  -H "Content-Type: application/json" \
  -d '{"plan_id":"<ID_DO_PLANO>"}'

# Acessar rota exclusiva para assinantes ativos
curl -X GET http://localhost:8080/api/v1/content/premium \
  -H "Authorization: Bearer <TOKEN_CLIENTE>"
```

---

## 🌿 Fluxo de Desenvolvimento (Git Workflow)

Este projeto foi construído utilizando **Feature Branches** (`--no-ff` merges) e **Conventional Commits**:

- `main`: Código estável e documentado.
- `chore/initial-setup`: Setup do módulo Go, Docker Compose, PostgreSQL, GORM e `/health`.
- `feat/users-auth`: Domínio de usuários, hash de senha (`bcrypt`) e autenticação JWT.
- `feat/plans-crud`: Domínio e endpoints de planos de assinatura com proteção RBAC (`ADMIN`).
- `feat/subscriptions`: Adesão, consulta e cancelamento de assinaturas + transações ACID no GORM.
- `feat/billing-worker`: Motor de faturamento recorrente (Goroutine + Graceful Shutdown) e middleware de assinante ativo.
- `test/unit-and-swagger`: Testes unitários nas camadas de `service`, `Dockerfile` multi-stage e guia de execução.
