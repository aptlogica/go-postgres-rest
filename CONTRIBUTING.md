# Contributing

Thank you for considering contributing to this project! Please follow these guidelines to help us maintain a high-quality, welcoming, and consistent open-source ecosystem.

## Code of Conduct

This project adheres to the Contributor Covenant Code of Conduct. By participating, you are expected to uphold this code. Violations should be reported to the maintainers.

## Development Setup

### Prerequisites

- Go 1.20 or later
- PostgreSQL 12 or later
- Docker and Docker Compose (optional, for containerized PostgreSQL)
- Git

### Local Environment Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/aptlogica/go-postgres-rest.git
   cd go-postgres-rest
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   go mod verify
   ```

3. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your local PostgreSQL credentials
   ```

4. **Run tests:**
   ```bash
   go test -v -race -coverprofile=coverage.out ./...
   ```

5. **View coverage:**
   ```bash
   go tool cover -html=coverage.out -o coverage.html
   open coverage.html
   ```

## Branch Naming Conventions

- **`feature/<name>`** - New features (e.g., `feature/add-query-builder`)
- **`fix/<name>`** - Bug fixes (e.g., `fix/connection-pool-leak`)
- **`docs/<name>`** - Documentation updates (e.g., `docs/update-api-reference`)
- **`refactor/<name>`** - Code refactoring (e.g., `refactor/simplify-repo-factory`)
- **`chore/<name>`** - Maintenance tasks (e.g., `chore/update-dependencies`)
- **`release/<version>`** - Release preparation (e.g., `release/v1.0.0`)

Branch names should be:
- Lowercase
- Hyphen-separated
- Descriptive but concise
- Limited to 50 characters

## Commit Conventions

This project follows [Conventional Commits](https://www.conventionalcommits.org/):

### Commit Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **`feat`** - New feature
- **`fix`** - Bug fix
- **`docs`** - Documentation changes
- **`style`** - Code style changes (formatting, missing semicolons, etc.)
- **`refactor`** - Code refactoring without feature changes
- **`perf`** - Performance improvements
- **`test`** - Adding or updating tests
- **`build`** - Build system or dependency changes
- **`ci`** - CI/CD pipeline changes
- **`chore`** - Maintenance tasks

### Scopes

Include the affected component:
- `postgres` - PostgreSQL database code
- `config` - Configuration module
- `models` - Data models
- `services` - Business logic services
- `utils` - Utility functions
- `examples` - Example applications

### Examples

```
feat(postgres): add bulk insert optimization

- Implement batch insert with configurable batch size
- Add BulkInsert method to repository interface
- Update performance benchmarks

Closes #42
```

```
fix(services): handle nil pointer on relationship sync

Previously failed when syncing empty relationships.
Now properly initializes relationship data.

Fixes #137
```

## Pull Request Checklist

Before submitting a pull request:

- [ ] Ensure your branch is based on `main` and up to date
- [ ] Code follows project style guidelines
- [ ] All tests pass: `go test -v -race ./...`
- [ ] Coverage maintained or improved: `go test -coverprofile=coverage.out ./...`
- [ ] Code linting passes: `go fmt ./...`
- [ ] Dead code removed: No unused imports or functions
- [ ] Documentation updated (README, docs/, code comments)
- [ ] Related issues referenced in PR description
- [ ] Commit messages follow Conventional Commits format

## Testing Requirements

### Code Coverage Standards

- **Overall project:** Minimum 90% coverage required for main release
- **Critical functions:** ParseValue, BuildComplexQuery, GetRelationshipData - minimum 85%
- **New code:** Must include tests with 80%+ coverage for new functions

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run specific test
go test -v ./pkg/database/postgres -run TestQueryBuilder

# Run benchmarks
go test -v -bench=. -benchmem ./...
```

### Test File Naming

- Test files: `*_test.go`
- Internal tests: `*_internal_test.go` (package-private testing)
- Location: Same directory as code being tested

## Documentation Requirements

### For New Features

- Add entry to CHANGELOG.md with feature description
- Update relevant docs in `docs/` directory
- Include example usage in code comments or docs/
- Update README if user-facing

### For API Changes

- Document changed signatures in migration guide
- Update API documentation in docs/
- Include deprecation notice if replacing existing API

### Code Comments

- Public functions: Brief description + parameter/return documentation
- Complex logic: Inline comments explaining non-obvious code
- TODO/FIXME: Mark with `// TODO:` or `// FIXME:` with issue reference

## Code Style Guide

- **Format:** Follow `go fmt` standards (enforced by CI)
- **Naming:** Use clear, descriptive names for variables and functions
- **Errors:** Always handle and propagate errors, never ignore
- **Logging:** Use structured logging for debugging
- **Security:** Never commit credentials, API keys, or sensitive data

## Review Process

1. **Create PR:** Push your branch and open a PR with detailed description
2. **Automated checks:** GitHub Actions must pass (tests, coverage, linting)
3. **Code review:** At least one maintainer reviews changes
4. **Changes requested:** Address feedback and push updates
5. **Approval:** Maintainer approves and merges to main

## Release Process

1. Update version in code and documentation
2. Update CHANGELOG.md with release notes
3. Create Git tag: `git tag -a vX.Y.Z -m "Release X.Y.Z"`
4. Push tag: `git push origin vX.Y.Z`
5. Create GitHub release: `gh release create vX.Y.Z --generate-notes`

See [VERSIONING.md](VERSIONING.md) for version numbering details.

## Getting Help

- **Questions:** Open a GitHub Discussion or Issue
- **Bugs:** Report with reproducible steps in a GitHub Issue
- **Security:** See SECURITY.md for reporting vulnerabilities
- **Chat:** Join our community discussions

## Attribution

Contributors will be recognized in:
- GitHub contributor graph
- Project README
- Release notes (for significant contributions)

Thank you for contributing!
