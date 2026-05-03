# Estratégia de Persistência - Divisão de Responsabilidades

Este documento define a separação de responsabilidades entre o **Conector de Banco de Dados** (Infraestrutura) e a **Camada de Repository** (Adapters), seguindo os princípios da Arquitetura Hexagonal adotada no projeto **Don**.

---

## 1. Conector de DB (Infraestrutura)
**Localização Sugerida:** `pkg/database`

Responsável pelo "como" a aplicação se conecta e mantém a integridade técnica da conexão com o banco de dados.

| Responsabilidade | Descrição |
| :--- | :--- |
| **Pool Management** | Configuração de limites de conexões (`MaxOpen`, `MaxIdle`, `Lifetime`). |
| **Health Check** | Implementação de Pings agressivos no startup para garantir disponibilidade. |
| **Observabilidade Global** | *Slow Query Logging* agnóstico e ganchos para Tracing (OpenTelemetry). |
| **Resiliência de Conexão** | Implementação de Circuit Breaker para evitar cascata de erros em falhas do banco. |
| **Migrations & Seeding** | Gerenciamento do ciclo de vida do esquema e dados iniciais (idempotência). |
| **Abstração de Transações** | Fornecer a infraestrutura técnica (ex: padrão *Atomic* via callback). |

---

## 2. Camada de Repository (Adapters)
**Localização Sugerida:** `internal/adapters/db`

Responsável pelo "o que" é feito com os dados, traduzindo as necessidades do domínio para a linguagem do banco (SQL).

| Responsabilidade | Descrição |
| :--- | :--- |
| **Resiliência de Query** | Uso obrigatório de `QueryContext` / `ExecContext`. |
| **Timeouts Específicos** | Definição de prazos máximos baseados na complexidade de cada consulta. |
| **Tratamento de Erros** | Tradução de erros de driver/infra para *Sentinel Errors* de domínio. |
| **Lógica de Transação** | Coordenação de quais operações devem ser atômicas (usando a infra do conector). |
| **Mapeamento (ORM/Manual)** | Conversão entre as entidades de domínio e as linhas da tabela. |

---

## 3. Coordenação (Core/Usecases)

*   **Unit of Work:** Os Casos de Uso devem coordenar transações através de interfaces, sem conhecer detalhes de implementação do driver SQL.
*   **Independência de Driver:** O Repository deve ocultar o driver específico (ex: PostgreSQL, MySQL) para que o domínio permaneça puro.
