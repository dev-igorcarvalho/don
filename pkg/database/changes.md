# Refactoring Plan: Database Package

This document outlines the required changes to `pkg/database` and its integration within `internal/adapters` to improve robustness, testability, and architectural integrity.

## Objectives

- **Decoupling**: Ensure `pkg/database` is a self-contained, reusable package.
- **Error Handling**: Standardize on returning raw errors to the caller, avoiding internal logging or panics.
- **Primary/Replica Support**: Introduce `SQLPair` for managing separate Reader and Writer connections.
- **Validation**: Ensure configuration is validated before attempting to open connections.
- **Integration**: Prepare `internal/adapters` to leverage the new structures.

## Required Changes

### 1. `pkg/database/sql.go`

- **Rename `NewSQL`**: Ensure consistent use of `NewSQL`.
- **Configuration**: 
    - Define a `Config` struct within `pkg/database`.
    - Fields: `Driver`, `DSN`, `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`, `ConnMaxIdleTime`, `Warmup`, `ConnectTimeout`.
- **Error Handling**:
    - Return errors directly from `NewSQL`.
    - NEVER panic or use `log.Fatal` inside this package.
- **Logging**: Remove dependency on `pkg/logger`.
- **SQLPair Structure**: 
    - Create a `SQLPair` struct:
      ```go
      type SQLPair struct {
          Writer *sql.DB
          Reader *sql.DB
      }
      ```
    - Implement `NewSQLPair(writerCfg, readerCfg Config) (*SQLPair, error)`.
    - Implement a `Close() error` method for `SQLPair` that closes both connections.

### 2. `internal/config/types.go`

- **Validation**: Add a `Validate() error` method to `DBConfig`.
- **Mapping**: Add a helper (e.g., `ToPkgConfig() database.Config`) to map `internal/config.DBConfig` to `pkg/database.Config`.

### 3. `internal/adapters/` (Integration)

- **`BaseRepository`**:
    - Update to accept `*database.SQLPair`.
    - Use `Writer` for mutations (`Exec`) and `Reader` for queries (`Query`, `QueryRow`).
- **`TransactionManager`**:
    - Ensure it uses the `Writer` from `SQLPair` for transactions.

### 4. `pkg/database/sql_test.go`

- Update tests to match new `Config` and constructor signatures.

## Implementation Plan

1.  **Phase 1: `pkg/database` Refactor**: [COMPLETED]
    - Define `database.Config`.
    - Refactor `NewSQL` to accept `Config` and remove logging.
    - Add `SQLPair` and `NewSQLPair`.
    - Update `sql_test.go`.
2.  **Phase 2: `internal/config` Enhancement**: [COMPLETED]
    - Add `Validate()` and mapping helper to `DBConfig`.
3.  **Phase 3: `internal/adapters` Integration**: [COMPLETED]
    - Update `BaseRepository` and `TransactionManager` to work with `SQLPair`.
4.  **Phase 4: Global Updates**: [COMPLETED]
    - Update `cmd/api/main.go` (if/when DB is wired) using `must.Do`.
5.  **Phase 5: Verification**: [COMPLETED]
    - Run `go test ./pkg/database/... ./internal/adapters/...`.
