# ADR-003: Factory Pattern for Repository Creation

## Status
Accepted

## Context

The project needs to instantiate repository objects with proper dependencies and configuration. As the number of repository types grows, creating them individually becomes error-prone and repetitive.

Options considered:
- Direct instantiation: Simple but couples code to concrete implementations
- Service Locator: Centralized lookup but can be hard to test
- Dependency Injection Container: Popular in Java/C#, complex for Go
- Factory Pattern: Simple, idiomatic Go approach

## Decision

We use the Factory Pattern to create repository instances. This provides:

- Centralized creation logic for repositories
- Easy to swap implementations for different databases
- Testable through mock factories
- Reduces boilerplate in consuming code

## Consequences

### Positive
- Single point to manage repository creation
- Easy to add new repositories
- Consumers don't need to know about dependencies
- Can easily inject mock factories in tests
- Clear separation of construction and usage

### Negative
- One more layer of indirection
- Factory must be kept in sync with repository interfaces
- Developers need to find and use factory instead of direct instantiation

## Implementation

- `pkg/database/factory.go` - Interface definitions
- `pkg/database/dbFactory.go` - Factory implementation
- `RepositoryFactory` interface defines all available repositories
- Factories are created per database connection
- Thread-safe factory implementations using sync.Once where appropriate
