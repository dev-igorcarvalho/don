# Infrastructure & Toolkit TODO

This document tracks the implementation of core infrastructure components and AWS adapters, following the Hexagonal Architecture and consumer-side interface standards.
## 1. Pilares Arquiteturais (Core & Estrutura)
- [x] **Injeção de Dependência Manual (Constructor Injection)**
  - [x] Definir contratos por interfaces no nível do consumidor (Consumer-side interfaces).
  - [x] Implementar o padrão *Manual Provider* (Container) no `main.go`.
- [x] **Provider de Configuração (Variáveis de Ambiente)**
  - [x] Criar struct centralizada para configuração (*Schema-First*).
  - [x] Implementar *Environment Hydration* com validação obrigatória no início (*Fail-fast*).
- [x] **Structured Logging Nativo (slog)**
  - [ ] Configurar handlers para JSON (prod) e Text (dev).
  - [x] Implementar *Contextual Logging* para propagação de `trace_id`.
- [x] **Graceful Shutdown & Context Propagation**
  - [x] Implementar escuta de sinais SIGINT e SIGTERM.
  - [x] Criar gerenciamento de ciclo de vida para fechamento ordenado de recursos.
- [x] **Estratégia de Erros**
  - [x] Definir *Domain Errors* constantes.
  - [x] Implementar *Opaque Errors* para logs detalhados e retornos amigáveis.
  - [x] Padronizar *Error Wrapping* com contexto em cada camada.
- [x] **Probes de Health & Readiness**
  - [x] Implementar verificação de integridade de dependências críticas (DB, etc.).
- [x] **Global Exception/Panic Recovery**
  - [x] Middleware para capturar panics e logar stack trace estruturado.

## 2. Camada de Transporte HTTP
- [x] **Decodificação Estrita de JSON**
  - [x] Usar `DisallowUnknownFields()`.
  - [x] Aplicar `http.MaxBytesReader` para limitar tamanho do body.
- [x] **Middleware de Contextualização**
  - [x] Geração/Propagação de `Request ID` via Header `X-Request-ID`.
  - [x] Aplicação de *Timeout* por requisição.
- [x] **Gestão de Headers**
  - [x] Validar `Content-Type: application/json`.
  - [x] Injetar *Security Headers* (nosniff, DENY, etc.).
- [x] **Padronização de Respostas de Erro**
  - [ ] Seguir o padrão RFC 7807 (Problem Details).
  - [x] Mapeamento automático de erros de domínio para status codes HTTP.
- [x] **Rate Limiting (Opt-in)**
  - [x] Implementar limitador simples por IP com headers de feedback.

## 3. Camada de Persistência (Database)
- [ ] **Pool Management & Fine-Tuning**
  - [x] Configurar limites explícitos de conexões (`MaxOpen`, `MaxIdle`, `Lifetime`).
  - [ ] Implementar Health Check/Ping agressivo no startup.
- [ ] **Abstração de Transações (Unit of Work)**
  - [ ] Implementar padrão *Atomic* via callback para evitar vazamento de `sql.Tx`.
  - [ ] (Opcional) Implementar transação via `context.Context`.
- [ ] **Resiliência de Operações**
  - [ ] Forçar uso de `QueryContext` / `ExecContext`.
  - [ ] Definir timeouts específicos por query.
  - [ ] Implementar Circuit Breaker básico para falhas consecutivas.
- [ ] **Observabilidade de Banco**
  - [ ] Implementar *Slow Query Logging* para consultas acima de um threshold.
  - [ ] Deixar gancho para Tracing Integrado.
- [ ] **Tratamento de Erros de Infra**
  - [ ] Mapear erros específicos do driver para erros de domínio (*Sentinel Errors*).
- [ ] **Migrations & Seeding**
  - [ ] Integrar execução de migrations ao ciclo de vida da aplicação.
  - [ ] Garantir idempotência em scripts de semente (*seeding*).

## 4. Funcionalidades Avançadas (Staff Level)
- [ ] **Paginação Baseada em Cursor**
  - [ ] Implementar lógica de `after_id` para evitar problemas de escala/concorrência.
- [ ] **Deep Health Checks**
  - [ ] Endpoint `/health/ready` com detalhes de uptime e versão (Commit Hash).

## 5. Extensibilidade e Configuração
- [ ] **Middlewares Componíveis**
  - [ ] Permitir ativação seletiva (Opt-in) por rota.
- [ ] **Contract Testing & Docs**
  - [ ] Toggle para habilitar/desabilitar documentação OpenAPI/Swagger.
- [ ] **Mock Mode**
  - [ ] Prover adapters de infraestrutura estáticos para desenvolvimento paralelo.

## 6. Explorar mais opcoes do linter com configs mais avançadas
- [ ] ** Validar funcoes avançadas de linter**
  - [ ] Descobrir as opcoes e implementar o que for pertinente

## 🌐 HTTP Client
- [ ] **Implement `pkg/httpclient`**
  - [ ] Resilient wrapper around `net/http`.
  - [ ] Support for middleware (logging, correlation IDs, retries).
  - [ ] Timeout management and context propagation.
- [ ] **Define Consumer-Side Interfaces**
  - [ ] Ports defined in `internal/core/usecases` (as needed).
- [ ] **Implement Adapters**
  - [ ] `internal/adapters/externalapi` for specific service integrations.

## ☁️ AWS Foundation
- [ ] **Implement `pkg/aws`**
  - [ ] AWS SDK v2 Configuration provider.
  - [ ] Support for environment-based configuration and LocalStack overrides.
  - [ ] Standardized client factory for DynamoDB, SQS, and SNS.

## 🗄️ DynamoDB (Database Adapter)
- [ ] **Implement `internal/adapters/aws/dynamodb`**
  - [ ] Repository implementations for domain entities.
  - [ ] Mapping logic between DynamoDB attributes and domain models.
  - [ ] Integration tests using LocalStack or DynamoDB Local.

## 📩 SQS (Messaging Adapter)
- [ ] **Implement `internal/adapters/aws/sqs`**
  - [ ] Producer implementation for pushing messages to queues.
  - [ ] Consumer/Worker logic (likely used in `cmd/worker`).
  - [ ] Dead-letter queue (DLQ) handling and retry strategies.

## 📢 SNS (Notification Adapter)
- [ ] **Implement `internal/adapters/aws/sns`**
  - [ ] Publisher implementation for domain events.
  - [ ] Topic management and message attribute filtering.

---
*Note: All implementations must include unit tests and follow the architectural directives in GEMINI.md.*
