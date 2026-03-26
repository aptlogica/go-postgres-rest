# ADR-004: Service Layer for Business Logic

## Status
Accepted

## Context

The application requires a layer to implement business logic, validation, and orchestration of repository operations. The separation between data access (repositories) and business logic (services) is important for:

- Maintainability and clarity of code organization
- Testability of business rules independent of data access
- Reusability of business logic across different interfaces (REST, GraphQL, CLI)
- Clear responsibility boundaries

Options considered:
- Business logic in HTTP handlers: Simple but mixes concerns
- Business logic in repositories: Violates single responsibility
- Dedicated service layer: Clean separation of concerns
- Domain-driven design with entities: More complex, fewer benefits for current project scope

## Decision

We implement a Service Layer pattern where:

- Repositories handle data persistence only
- Services contain all business logic and validation
- Services may use multiple repositories
- Services are used by all interface layers

## Consequences

### Positive
- Clear separation between data and business logic
- Business logic can be tested without database
- Service methods define available use cases clearly
- Easier to add new interface layers (GraphQL, gRPC, etc.)
- Business rules are centralized and easier to maintain

### Negative
- More files and interfaces to maintain
- Additional abstraction layer
- Potential overhead from extra function calls

## Implementation

- Services defined in `pkg/services/interfaces/`
- Implementations in `pkg/services/{entity}_service.go`
- All services use constructor injection for repositories
- Services handle validation, error mapping, and orchestration
- Services follow naming: `{Entity}Service` interface, `postgres{Entity}Service` for implementations
