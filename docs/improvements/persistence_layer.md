# 3. Camada de Persistência (Database)

Este documento detalha as melhorias planejadas para a camada de persistência, focando em resiliência, performance e manutenibilidade.

## 3.1. Pool Management & Fine-Tuning

### Motivo
O comportamento padrão do `sql.DB` pode levar à exaustão de recursos no banco de dados ou no sistema operacional se não for configurado. Limites explícitos garantem previsibilidade sob carga.

### Implementação

```go
func setupDB(cfg config.DBConfig) *sql.DB {
    db, err := sql.Open("postgres", cfg.DSN)
    if err != nil {
        log.Fatal(err)
    }

    // Limites explícitos para evitar vazamento de recursos
    db.SetMaxOpenConns(cfg.MaxOpenConns) // Máximo de conexões simultâneas
    db.SetMaxIdleConns(cfg.MaxIdleConns) // Máximo de conexões ociosas no pool
    db.SetConnMaxLifetime(time.Hour)    // Tempo máximo de vida de uma conexão

    // Health Check agressivo no startup
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        log.Fatalf("falha ao conectar no banco: %v", err)
    }

    return db
}
```

**Por que?**
- `MaxOpenConns`: Evita o erro "too many connections" no Postgres/MySQL.
- `MaxIdleConns`: Mantém um pool "quente" sem manter conexões desnecessárias abertas para sempre.
- `PingContext`: Garante que a aplicação só suba se a infraestrutura estiver pronta.

---

## 3.2. Abstração de Transações (Unit of Work)

### Motivo
Gerenciar `sql.Tx` manualmente frequentemente leva a esquecimentos de `Commit` ou `Rollback`, além de poluir a camada de Use Case com detalhes de infraestrutura.

### Implementação (Padrão Atomic via Callback)

```go
type TransactionManager interface {
    Atomic(ctx context.Context, fn func(ctx context.Context) error) error
}

func (r *repository) Atomic(ctx context.Context, fn func(ctx context.Context) error) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    // Injetar o TX no contexto ou passar por parâmetro (preferencialmente contexto para transparência)
    ctxWithTx := context.WithValue(ctx, txKey{}, tx)

    err = fn(ctxWithTx)
    if err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("rollback error: %v (original error: %w)", rbErr, err)
        }
        return err
    }

    return tx.Commit()
}
```

**Por que?**
- O callback garante que o ciclo de vida da transação seja fechado corretamente.
- Remove a responsabilidade de `Commit/Rollback` do desenvolvedor do Use Case.

---

## 3.3. Resiliência de Operações

### Motivo
Operações de banco de dados podem travar ou falhar de forma intermitente. Precisamos de timeouts e proteção contra falhas em cascata.

### Implementação

```go
func (r *repository) GetUser(ctx context.Context, id string) (*User, error) {
    // Forçar uso de Context para cancelamento e timeouts
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    var user User
    query := "SELECT id, name FROM users WHERE id = $1"
    
    // Uso obrigatório de QueryContext/ExecContext
    err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name)
    if err != nil {
        return nil, r.mapError(err)
    }

    return &user, nil
}
```

**Por que?**
- `QueryContext`: Permite que o banco interrompa a execução se o cliente (API) desconectar.
- `Circuit Breaker`: (Opcional) Protege o banco de ser bombardeado quando já está degradado.

---

## 3.4. Observabilidade de Banco

### Motivo
"Queries lentas" são a causa número 1 de degradação de performance. Precisamos de visibilidade sem instrumentação manual em cada método.

### Implementação (Decorator ou Middleware)

```go
func (r *repository) queryWithLogging(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    start := time.Now()
    rows, err := r.db.QueryContext(ctx, query, args...)
    duration := time.Since(start)

    if duration > slowQueryThreshold {
        slog.WarnContext(ctx, "slow query detected", 
            "query", query, 
            "duration", duration,
            "args", args)
    }

    return rows, err
}
```

**Por que?**
- Identifica gargalos de performance em produção proativamente.
- Facilita a correlação entre tempo de resposta e consultas específicas.

---

## 3.5. Tratamento de Erros de Infra

### Motivo
A camada de domínio não deve saber se estamos usando Postgres, MySQL ou um arquivo CSV. Erros como `sql.ErrNoRows` ou códigos do driver não devem vazar.

### Implementação (Sentinel Errors)

```go
var (
    ErrNotFound = errors.New("resource not found")
    ErrConflict = errors.New("resource already exists")
)

func (r *repository) mapError(err error) error {
    if errors.Is(err, sql.ErrNoRows) {
        return ErrNotFound
    }
    
    // Mapeamento específico de driver (ex: Postgres code 23505 para Unique Violation)
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23505" {
        return ErrConflict
    }

    return err
}
```

---

## 3.6. Migrations & Seeding

### Motivo
Garantir que todos os ambientes (Dev, CI, Prod) estejam com o schema sincronizado de forma automática e segura.

### Estratégia
- **Migrations**: Integrar o `golang-migrate` ou `goose` no startup da aplicação (ou via CLI sidecar).
- **Seeding**: Scripts de carga inicial devem ser idempotentes (`INSERT ... ON CONFLICT DO NOTHING`).

```go
func runMigrations(db *sql.DB) error {
    // Lógica para aplicar arquivos .sql ou migrations embarcadas
    // Garante que o banco está na versão correida antes de aceitar tráfego
}
```
