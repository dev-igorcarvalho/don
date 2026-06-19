I have analyzed the pkg/server/echoserver package and compared it with the project's
structure and standards. Here is a summary of my findings and suggestions:

Current State
* Good Abstraction: The package provides a clean wrapper around the Echo framework,
  using functional options (New(opts ...Option)) and a struct-based routing system
  (Route, Group).
* Interface-Driven: Enforcing a Handler interface (Handle(c echo.Context) error) is a
  solid practice that facilitates testing and dependency injection.
* Test Coverage: Basic functionality (initialization, route registration, group
  registration, and shutdown) is well-covered with tests using httptest.

Missing Features & Suggestions

1. Logging Integration:
    * Observation: You have a pkg/logger package, but the server does not currently
      integrate with it.
    * Suggestion: Add a WithLogger functional option to EchoServer and implement a
      middleware that uses your custom logger for request logging.

2. Standardized Error Handling:
    * Observation: The server currently uses Echo's default error handler.
    * Suggestion: Implement a custom global error handler to ensure all API errors
      return a consistent JSON response (e.g., { "error": "message", "code": 400 }).

3. Recovery Middleware:
    * Observation: There's no built-in protection against panics.
    * Suggestion: Add Echo's middleware.Recover() by default or as a standard option to
      prevent the server from crashing on unexpected errors.

4. Request Validation:
    * Observation: There is no centralized way to handle request body validation.
    * Suggestion: Integrate a validator (like go-playground/validator) by setting
      s.app.Validator.

5. Graceful Shutdown Signal Handling:
    * Observation: While a Shutdown method exists, the logic for listening to OS signals
      (SIGINT, SIGTERM) is likely left to main.go.
    * Suggestion: Add a helper method (e.g., Wait(signals ...os.Signal)) that blocks
      until a signal is received, making main.go cleaner.
    * Explanation: Encapsulating signal handling within the server abstraction simplifies 
      the application's entry point. It removes boilerplate code from `main.go` and 
      provides a standard way for the process to block until a termination signal is 
      received, facilitating a clean transition to the shutdown sequence.

6. Readiness/Liveness Probes:
    * Observation: You have a health handler in internal/handlers/health.go, but it must
      be manually registered.
    * Suggestion: Provide a WithHealthCheck(path string) option that automatically
      registers a basic health route.