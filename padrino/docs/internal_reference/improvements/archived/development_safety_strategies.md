# Development Safety & Integrity Strategies

This document outlines the strategies and guardrails designed to ensure that development remains surgical, intentional, and free of unintended side effects.

---

## 1. The "Contract Gatekeeper" (Schema Validation)
Instead of relying on auto-generated code, we use **JSON Schemas** to validate that API responses do not drift from the agreed-upon contract.

### The "Contract Gatekeeper" (Manual but Strict)
Instead of auto-generating code, we use JSON Schema or Handwritten Spec Files as a validation layer in your tests.
* The Idea: We create a directory tests/contracts/ containing .json files representing the expected response for every endpoint.
* The Guardrail: I am forbidden from finishing a task until I run a test that validates the current Handler output against that JSON schema.
* Why: This prevents me from accidentally renaming a field from user_id to userID (a classic "fuck the code" moment) because the schema validation will fail.

### Implementation Example:
**File:** `internal/handlers/testdata/health_schema.json`
```json
{
  "type": "object",
  "properties": {
    "status": { "type": "string", "enum": ["up", "down"] }
  },
  "required": ["status"]
}
```

**File:** `internal/handlers/health_contract_test.go`
```go
func TestHealthContract(t *testing.T) {
    rec := httptest.NewRecorder()
    // ... call handler ...
    
    schema := loadSchema("testdata/health_schema.json")
    result := validateJSON(rec.Body.Bytes(), schema)
    
    assert.True(t, result.Valid(), "Response drifted from contract!")
}
```

---

## 2. Skill: `test-scaffolder` (Logic Mapping)
A workflow that forces the identification of all logical branches before implementation begins.

### Skill: test-scaffolder (Logic Mapping)
Your idea of scanning for 100% coverage targets is excellent. Here is how that workflow would look:
* Phase 1 (Analysis): I scan the Go file and output a "Logic Map" (e.g., "Function X has 3 branches: success, invalid ID, DB error").
* Phase 2 (Scaffolding): I create _test.go with empty Test... functions for each branch, all containing t.Skip("Not implemented").
* Phase 3 (Implementation): Only after you approve the "Logic Map," I implement the tests.
* Rules: 
  * I cannot modify the implementation code until the test scaffolds are in place and failing (TDD approach).
  * Never edit the implementation code just to make a test pass, if a change make a plan and ask the user for permissions

### Implementation Example:
**Command:** `gemini activate_skill test-scaffolder && gemini run "scan internal/core/usecases/user.go"`

**Output (The Logic Map):**
| Function | Branch | Condition | Expected Result |
| :--- | :--- | :--- | :--- |
| `CreateUser` | Success | Valid input, Email unique | Return User, nil error |
| `CreateUser` | Duplicate | Email already exists | Return nil, ErrDuplicate |
| `CreateUser` | Validation | Empty name | Return nil, ErrInvalidInput |

**Result:** Generates `user_test.go` with 3 empty `t.Run` blocks. Implementation only starts after user signs off on the map.

---

## 3. Architectural Guardrails (Dependency Linting)
Enforces the **Hexagonal Architecture** by preventing illegal imports (e.g., Domain importing Adapters).

###  Architectural Guardrails (Linter-as-Policy)
Since you are using Hexagonal Architecture, the biggest "fuck up" I can make is leaking concerns (e.g., putting an SQL query inside a Use Case).
* Suggestion: We implement go-arch-lint (https://github.com/fe3dback/go-arch-lint). It allows us to define a dep-rules.yml that says: "Internal/Core cannot import Internal/Adapters."
* The Guardrail: Every time I make a change, I must run the arch-linter. If I try to take a shortcut that breaks the architecture, the CI/Linter catches it immediately.

### Implementation Example:
**File:** `.go-arch-lint.yml`
```yaml
commonComponents:
  - internal/core/domain
  - internal/core/usecases

check:
  internal/core/domain:
    mayNotImport:
      - internal/adapters/**
      - internal/handlers/**
  internal/core/usecases:
    mayNotImport:
      - internal/adapters/**
```

**Guardrail:** `go-arch-lint check` must be run after every modification to `internal/`.

---

## 4. Impact Analysis (Pre-Flight Check)
A mandatory research phase to map the "Blast Radius" of a change.

###  Impact Analysis (Pre-Flight Check)
Before I touch any code in internal/core/domain, I should be forced to perform an Impact Analysis.
* The Requirement: I must search the entire project for every file that imports the struct I'm about to change.
* The Output: I present you with a list: "Changing User struct affects 4 handlers and 2 use cases. Here is the migration plan."
* Why: It stops me from making "local" changes that have global side effects.

### Implementation Example:
**Manual Step for Agent:**
1. User asks to change `domain.User` struct.
2. Agent runs: `grep -r "domain.User" ./internal`.
3. Agent reports:
   > **Impact Analysis:**
   > - `internal/handlers/user_handler.go`: Needs update to JSON tags.
   > - `internal/adapters/db/user_repository.go`: Needs SQL query update.
   > - `internal/core/usecases/register_user.go`: Needs logic adjustment.
   > **Proceed with changes to these 3 files?**

---

## 5. Snapshot/Golden File Testing
Ensures that refactoring complex logic results in byte-for-byte identical output.

###  Snapshot Testing for Complex Logic
For deep business logic, we can use Golden Files (Snapshot testing).
* The Idea: If I'm refactoring a complex calculator or data processor, we run the code once to save the output to a testdata/output.golden file.
* The Guardrail: Any change I make must result in the exact same "Golden" output. If the byte-by-byte output changes, I have to explain to you why it changed before you accept the PR.

### Implementation Example:
**File:** `internal/core/usecases/report_generator_test.go`
```go
func TestReportGenerator_Golden(t *testing.T) {
    actual := generator.Generate(inputData)
    
    if *update {
        os.WriteFile("testdata/report.golden", actual, 0644)
    }

    expected, _ := os.ReadFile("testdata/report.golden")
    assert.Equal(t, expected, actual, "Refactor changed the output format!")
}
```
*Note: This is critical for data-heavy operations where a small logic change could break downstream systems.*

---

## 4. Performance Regression Guardrails (Benchmarking)
Ensures that new code or refactors do not negatively impact the application's performance.

### The "Benchmark Baseline" (Performance Monitoring)
Instead of guessing if a change is efficient, we use Go's built-in benchmarking tools to compare new implementations against a known baseline.
* **The Idea:** For performance-critical code (e.g., hot paths in use cases or shared utility packages), we maintain benchmarks in `_test.go` files.
* **The Guardrail:** Before finalizing a change in a performance-sensitive area, I must run `go test -bench .` and compare it with the current baseline performance.
* **Why:** This prevents "death by a thousand cuts" where small, unvetted changes gradually slow down the system.

### Implementation Example:
**File:** `pkg/json/json_test.go`
```go
func BenchmarkMarshalLargeStruct(b *testing.B) {
    data := generateLargeStruct()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = json.Marshal(data)
    }
}
```

**Workflow:**
1. Run benchmarks on the current (modified) code: `go test -bench . -count 5 > new_bench.txt`
2. Compare results against the established baseline.
3. If a significant regression is found, I must report it and optimize before proceeding.
