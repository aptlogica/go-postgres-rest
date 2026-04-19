# Project Governance

This document describes the governance structure and decision-making process for the go-postgres-rest project.

## Overview

go-postgres-rest is an open-source project maintained by [Aptlogica Technologies](https://www.aptlogica.com) under the Apache License 2.0. We welcome contributions from the community and follow transparent governance practices.

## Project Roles

### Maintainers

**Primary Maintainers:**
- [@gauravgaikwad](https://github.com/gauravgaikwad) - Architecture, Core Database Layer

**Responsibilities:**
- Review and merge pull requests
- Manage releases and version control
- Set project direction and priorities
- Maintain code quality standards
- Address security concerns
- Communicate with community

### Contributors

**Types of Contributors:**
- Code contributors (bug fixes, features)
- Documentation contributors
- Reporting issues and bugs
- Providing feedback and suggestions

**How to Contribute:**
See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed guidelines.

## Decision-Making Process

### Minor Changes (Bug Fixes, Documentation)
- Approval by 1 maintainer required
- Can be merged immediately after approval
- Issue-to-PR traceability expected

### Feature Additions
- Proposal discussion in issue or RFC (see below)
- Approval by lead maintainer required
- Code review by 1+ maintainers
- Tests and documentation required

### Breaking Changes (Major Version)
- Community discussion required
- Clear deprecation path in previous version
- Documentation of migration guide
- 2-week notice before removal

### Security Issues
- Handled confidentially (see [SECURITY.md](../SECURITY.md))
- Fixed in dedicated branch
- Released as patch version
- Disclosure coordinated with community

## RFC (Request for Comments) Process

For major features or significant architectural changes:

1. **Create Discussion Issue:**
   - Labeled: `type/rfc`
   - Title: RFC: [Feature Name]
   - Include: motivation, implementation approach, alternatives

2. **Community Feedback (2-3 weeks):**
   - Maintainers and community provide feedback
   - Author refines proposal based on feedback

3. **Decision:**
   - Lead maintainer makes final decision
   - Rationale documented in issue

4. **Implementation:**
   - Create feature branch
   - Reference RFC in PR
   - Follow standard code review process

## Release Process

### Semantic Versioning

go-postgres-rest follows [Semantic Versioning](../VERSIONING.md):

- **Major (X.0.0):** Breaking API changes
- **Minor (X.Y.0):** New features, backward compatible
- **Patch (X.Y.Z):** Bug fixes

### Release Cycle

- **Maintenance Release (Patch):** As needed for critical fixes
- **Feature Release (Minor):** Monthly, around month-end
- **Major Release:** 1-2 per year, planned in roadmap

### Release Steps

1. Update version in code/docs
2. Update CHANGELOG.md
3. Create git tag: `vX.Y.Z`
4. Push to GitHub: `git push origin vX.Y.Z`
5. Automated release workflow creates GitHub release
6. Docker images published automatically
7. Announce via social channels

## Code Review Standards

### Requirements for All PRs

- Passes all CI checks (tests, linting, security)
- Test coverage maintained or improved (>87%)
- Documentation updated if needed
- No breaking changes without discussion

### Review Checklist

- [ ] Code follows project style guide
- [ ] Tests included and passing
- [ ] Comments explain non-obvious logic
- [ ] Error handling is appropriate
- [ ] Documentation is updated
- [ ] No security vulnerabilities introduced

## Roadmap

The project roadmap is maintained in GitHub Issues and Projects:

- **Current Release**: Fixes and small features
- **Next Release (1-2 months)**: Planned features
- **Future (6+ months)**: Vision items

Contributors can:
- View roadmap to align work
- Propose changes via issues
- Vote on feature priorities

## Communication

### Primary Channels

- **Issues:** Bug reports, feature requests, questions
- **Discussions:** RFC, design decisions, announcements
- **Email:** support@aptlogica.com (security only)

### Response Times

- **Critical Bugs:** <24 hours
- **Other Issues:** <1 week
- **PRs:** <2 weeks for initial review

## Moderation Policy

This project maintains a welcoming environment. See [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md).

Violations may result in:
- Warning
- Comment removal
- Temporary muting
- Permanent ban (severe cases)

## Stability & Deprecation

### API Stability Guarantee

- **v0.x:** API may change (breaking changes possible)
- **v1.0+:** API stability guaranteed for patch versions

### Deprecation Policy

- Features marked deprecated for at least 2 major versions
- Clear migration path provided
- Warnings logged when deprecated feature used

## Contributing to Governance

We welcome feedback on governance process:

1. Open issue with label `governance`
2. Describe proposed change and rationale
3. Discuss with maintainers
4. If approved, update this document
5. Announce change to community

## FAQ

**Q: How do I become a maintainer?**
A: Demonstrate consistent, high-quality contributions over time. Discuss with existing maintainers.

**Q: Can I fork and maintain my own version?**
A: Yes! Apache License 2.0 permits this. We hope you'll contribute improvements back.

**Q: How are security issues handled?**
A: See [SECURITY.md](../SECURITY.md) for confidential reporting process.

**Q: What if I disagree with a decision?**
A: Open an issue to discuss. Maintainers will reconsider if new information provided.

---

**Last Updated:** March 25, 2026  
**Version:** 1.0
