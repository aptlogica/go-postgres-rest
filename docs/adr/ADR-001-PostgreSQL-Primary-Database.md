# ADR-001: Use PostgreSQL as Primary Database

## Status
Accepted

## Context

The project requires a robust, scalable relational database for managing complex data structures and relationships. The choice of database is fundamental to the project's architecture and long-term viability.

Multiple options were evaluated:
- PostgreSQL: Open-source, feature-rich, proven in production
- MySQL: Wide adoption, good compatibility
- NoSQL databases: Document-oriented, better for unstructured data
- Cloud databases: Managed services, reduced operational overhead

## Decision

We have decided to use PostgreSQL as the primary database for go-postgres-rest. PostgreSQL was selected as the database technology because:

1. **Advanced Features**: Full-text search, JSON support, arrays, and custom types
2. **ACID Compliance**: Strong consistency guarantees for financial and critical data
3. **Scalability**: Native support for partitioning and replication
4. **Community**: Large, active community with excellent documentation
5. **Cost**: Open-source and free, no licensing fees
6. **Integration**: Go has excellent PostgreSQL drivers (lib/pq)

## Consequences

### Positive
- Robust transaction support for complex operations
- Strong type system and schema validation
- Advanced indexing strategies for performance optimization
- Easy integration with Go using standard drivers
- Wide industry adoption and skill availability

### Negative
- Additional operational overhead for database administration
- Need to manage PostgreSQL installation and maintenance
- Requires SQL knowledge for optimal query design
- Vertical scaling limitations (horizontal scaling requires additional tooling)

### Mitigations
- Use containerization (Docker) for simplified deployment
- Implement connection pooling to manage resources
- Provide clear migration and backup procedures
- Document common performance tuning techniques

## Implementation

- PostgreSQL 12+ is the minimum supported version
- Use lib/pq Go driver for database connectivity
- Implement connection pooling with configurable pool size
- Support for both local and cloud-hosted PostgreSQL instances
