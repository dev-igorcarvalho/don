# SQL Adapter & Transaction Management Strategies

## Objective
Design a high-quality, production-ready SQL adapter for the `don` project (Hexagonal Architecture). The adapter will wrap `pkg/database.SQLPair` to support primary/replica routing and provide robust transaction management. As requested, this plan outlines industry-standard (Uber/Hashicorp inspired) approaches to handle database transactions within the repository layer, strictly adhering to consumer-side interface principles.

## Key Context & Constraints
*   **Architecture:** Hexagonal (Ports and Adapters). Repositories sit in `internal/adapters` and implement interfaces defined in `internal/core/usecases` (or `domain`).
*   **Base Connection:** `pkg/database.SQLPair` provides `.Writer()` and `.Reader()` which return `*sql.DB`.
*   **Go Standards:** Strict adherence to Go conventions (**Consumer-Side Interfaces**, no global state, explicit errors, dependency injection).

---

## The Approaches

### Approach 1: Explicit DB Interface (Consumer-Side)
In this approach, the repository defines exactly what it needs from a database connection (or transaction) as an interface. The use-case (the consumer) manages the transaction lifecycle and provides the repository with a concrete implementation that satisfies that interface.

**Code Example:**

```go
// internal/adapters/user_repo.go

// UserRepository defines what it needs from the DB/Tx.
// This interface is defined by the consumer (the repository in this case) 
// to decouple itself from the concrete sql.DB or sql.Tx.
type sqlQuerier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type UserRepository struct {
	db sqlQuerier
}

func NewUserRepository(db sqlQuerier) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.ExecContext(ctx, "INSERT ...")
	return err
}

// internal/core/usecases/user_usecase.go

// UserRepo is the interface defined by the UseCase (the consumer).
type UserRepo interface {
	Create(ctx context.Context, user *domain.User) error
}

// Transactor is the interface for managing transactions, defined by the UseCase.
type Transactor interface {
	RunInTx(ctx context.Context, fn func(q sqlQuerier) error) error
}

func (uc *UserUseCase) Register(ctx context.Context, user *domain.User) error {
	return uc.transactor.RunInTx(ctx, func(q sqlQuerier) error {
		// Instantiate a short-lived repo for this transaction
		// Note: sqlQuerier here would also be defined/shared appropriately
		repo := adapters.NewUserRepository(q)
		return repo.Create(ctx, user)
	})
}
```

**Advantages & Trade-offs:**
*   **PRO - Idiomatic Go:** Follows the "Accept interfaces, return structs" and "Consumer-side interfaces" principles perfectly.
*   **PRO - Highly Explicit:** Dependencies are clear. No hidden context magic.
*   **PRO - Easy to Mock:** The interface is focused and easy to mock for unit tests.
*   **CON - Allocation Overhead:** Requires instantiating new repository structs if switching between DB and Tx modes.

---

### Approach 2: Context-based Transactions (Middleware Strategy)
Transactions are managed by a "Transactor" and stored in the `context.Context`. The repository (the consumer of the context) extracts the transaction if present, otherwise uses the standard connection.

**Code Example:**

```go
// internal/adapters/sqladapter/context.go

// Note: No public interfaces defined here. 
// Only internal helpers for context management.

type txKey struct{}

func InjectTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// internal/adapters/user_repo.go

type UserRepository struct {
	db *database.SQLPair
}

// getQuerier is an internal helper that chooses between Tx in context or the base DB.
func (r *UserRepository) getQuerier(ctx context.Context) sqlQuerier {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return r.db.Writer()
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	q := r.getQuerier(ctx)
	_, err := q.ExecContext(ctx, "INSERT ...")
	return err
}
```

**Advantages & Trade-offs:**
*   **PRO - Clean API:** Use-cases don't need to manually pass DB/Tx objects around.
*   **PRO - Standardized:** Widely used in large-scale Go projects (e.g., Uber).
*   **CON - Hidden Dependencies:** The dependency on a transaction is "hidden" in the context, making it harder to trace without looking at the implementation.
*   **CON - Context Pollution:** Uses the context for control flow, which is sometimes debated in the Go community.

---

## Conclusion
Both approaches now strictly follow the **Consumer-Side Interface** requirement. Approach 1 is more explicit and favored for its clarity, while Approach 2 offers a more ergonomic API for complex business logic involving multiple repositories.
