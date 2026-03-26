# ADR-005: Error Handling and Recovery Strategy

## Status
Accepted

## Context

The project needs a consistent approach to error handling that:

- Preserves error context and stack information for debugging
- Provides meaningful error messages to clients
- Allows different layers to handle errors appropriately
- Distinguishes between recoverable and non-recoverable errors

Options considered:
- Ignore errors: Dangerous, leads to silent failures
- Log all errors: Loses context in deeply nested calls
- panic/recover: Premature, Go standard is error returns
- Wrapped errors (Go 1.13+): Standard Go approach with context
- Custom error types: Allows typed error handling

## Decision

We use Go's standard error handling with wrapped errors:

1. All functions return error as second return value
2. Errors are wrapped at origin with context: `fmt.Errorf("operation failed: %w", err)`
3. Each layer may wrap errors with additional context
4. Callers check errors immediately
5. Clients receive structured error responses

## Consequences

### Positive
- Idiomatic Go error handling
- Full error chain preserved for debugging
- Errors can be examined at different layers
- Works well with logging and monitoring
- Standard tools like errors.Is() and errors.As() available

### Negative
- Requires discipline to wrap errors consistently
- Verbose error checking code
- Client must parse error messages for specific cases

## Implementation

- Functions return error as last return value
- Errors wrapped with context at each layer
- Services map repository errors to domain errors
- HTTP handlers map domain errors to HTTP status codes
- Panic only for truly unrecoverable situations
- Structured logging captures full error context
- Error handling tested alongside happy paths
