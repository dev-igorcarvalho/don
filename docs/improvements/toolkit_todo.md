# Toolkit para Entrevista de Engenharia de Software (Go) - To-Do List

Este documento consolida os requisitos e padrões discutidos para a criação de um toolkit de alto nível, focado em arquitetura limpa e robustez sem dependências externas desnecessárias.

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
- [ ] **Decodificação Estrita de JSON**
    - [ ] Usar `DisallowUnknownFields()`.
    - [ ] Aplicar `http.MaxBytesReader` para limitar tamanho do body.
- [ ] **Validação Dupla**
    - [ ] Implementar validação sintática via struct tags.
    - [ ] Implementar método `Validate()` semântico para regras de negócio complexas.
- [ ] **Middleware de Contextualização**
    - [ ] Geração/Propagação de `Request ID` via Header `X-Request-ID`.
    - [ ] Aplicação de *Timeout* por requisição.
- [ ] **Gestão de Headers**
    - [ ] Validar `Content-Type: application/json`.
    - [ ] Injetar *Security Headers* (nosniff, DENY, etc.).
- [ ] **Padronização de Respostas de Erro**
    - [ ] Seguir o padrão RFC 7807 (Problem Details).
    - [ ] Mapeamento automático de erros de domínio para status codes HTTP.
- [ ] **Rate Limiting (Opt-in)**
    - [ ] Implementar limitador simples por IP com headers de feedback.

## 3. Camada de Persistência (Database)
- [ ] **Pool Management & Fine-Tuning**
    - [ ] Configurar limites explícitos de conexões (`MaxOpen`, `MaxIdle`, `Lifetime`).
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
- [ ] **Idempotência Nativa**
    - [ ] Middleware para Header `X-Idempotency-Key` com suporte a cache/banco.
- [ ] **Functional Options Pattern**
    - [ ] Usar para construtores de servidores, loggers e clientes.
- [ ] **Paginação Baseada em Cursor**
    - [ ] Implementar lógica de `after_id` para evitar problemas de escala/concorrência.
- [ ] **Deep Health Checks**
    - [ ] Endpoint `/health/ready` com detalhes de uptime e versão (Commit Hash).

## 5. Extensibilidade e Configuração
- [ ] **Middlewares Componíveis**
    - [ ] Permitir ativação seletiva (Opt-in) por rota.
- [ ] **Feature Flags Internas**
    - [ ] Implementar *Debug Mode* para logs de trace/body.
    - [ ] Implementar *Maintenance Mode* (503) controlado via config.
- [ ] **Contract Testing & Docs**
    - [ ] Toggle para habilitar/desabilitar documentação OpenAPI/Swagger.
- [ ] **Mock Mode**
    - [ ] Prover adapters de infraestrutura estáticos para desenvolvimento paralelo.
