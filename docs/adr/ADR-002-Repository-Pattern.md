# ADR-002: Repository Pattern for Data Access

## Status
Accepted

## Context

The application needs a data access layer to abstract database operations and provide a clean separation between business logic and persistence. This abstraction allows for:

- Testing business logic without database dependencies
- Switching database implementations if needed
- Consistent error handling across the application
- Deferred database operation execution

Multiple patterns were considered:
- Repository Pattern: Abstraction over data persistence
- DAO (Data Access Object): Similar to repository, used in Java
- Query Builder Pattern: Direct database query construction
- ORM (Object-Relational Mapping): Full object persistence layer

## Decision

We have adopted the Repository Pattern for all data access operations in the go-postgres-rest project.

The Repository Pattern provides:
- A collection-like abstraction for data access
- Separation of concerns between domain logic and persistence
- Testability through mock repositories
- Consistency across different entity types

## Consequences

### Positive
- Business logic is decoupled from database implementation
- Easy to write unit tests by mocking repositories
- Consistent interface for all data operations
- Can switch database backends with implementation changes only
- Clear contract for what operations are available

### Negative
- Additional abstraction layer adds complexity
- More code to maintain for repository implementations
- Potential performance impact from abstraction (mitigated by Go's inlining)
- Requires discipline to avoid breaking repository encapsulation

## Implementation

- All database operations go through repository interfaces
- Repositories defined in `pkg/database/interfaces/`
- PostgreSQL implementation in `pkg/database/postgres/`
- Factory pattern used to create repository instances
- Repositories follow naming convention: `{Entity}Repository` interface
- Implementations: `postgres{Entity}Repository` struct
