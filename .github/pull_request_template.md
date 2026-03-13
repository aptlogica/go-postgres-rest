# Pull Request

## Description

Brief description of the changes in this PR.

## Type of Change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Code refactoring
- [ ] Performance improvement

## Related Issues

Closes #[issue number]

## Changes Made

- [ ] Change 1
- [ ] Change 2
- [ ] Change 3

## Testing

### Testing Done

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed
- [ ] All existing tests pass

### Test Coverage

- [ ] Test coverage maintained or improved
- [ ] New code has appropriate test coverage

### Database Testing

- [ ] Migration tested (if applicable)
- [ ] Schema changes validated
- [ ] Performance impact assessed
- [ ] Tested with multiple PostgreSQL versions

### Testing Instructions

Provide step-by-step instructions for reviewers to test your changes:

1. Set up PostgreSQL with: `docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=test postgres:15`
2. Run migrations: `go run ./cmd/migrate up`
3. Start server: `go run ./cmd/server`
4. Test endpoint: `curl http://localhost:8080/api/...`

## Code Quality

- [ ] Code follows Go best practices and project style guidelines
- [ ] Self-review of code completed
- [ ] Code is properly documented with Go comments
- [ ] No debug code left in (fmt.Print, log statements, etc.)
- [ ] Error handling is comprehensive
- [ ] Context is properly handled for cancellation

## Breaking Changes

- [ ] No breaking changes
- [ ] Breaking changes documented in CHANGELOG.md
- [ ] Migration guide provided (if applicable)
- [ ] Database schema changes documented

## Performance Impact

- [ ] No performance impact
- [ ] Performance improvement (include benchmarks)
- [ ] Potential performance regression (documented and justified)

## Reviewers

@mention specific people you want to review this PR

## Additional Notes

Any additional information that would be helpful for reviewers.