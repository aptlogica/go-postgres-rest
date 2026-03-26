# Versioning

This project follows [Semantic Versioning](https://semver.org/) (SemVer) for release numbering and version management.

## Version Format

Versions are numbered using the format `X.Y.Z`, where:

- **X** = Major version (breaking changes, significant API modifications)
- **Y** = Minor version (new features, backward compatible additions)
- **Z** = Patch version (bug fixes, performance improvements, backward compatible)

Example: `v1.2.3`

### Pre-release Versions

Pre-release versions are denoted by appending a hyphen and identifiers: `v1.0.0-beta`, `v1.0.0-rc.1`

### Build Metadata

Build metadata can be appended with a plus sign: `v1.0.0+build.20250325`

## Release Schedule

- **Major versions** (X.0.0): Released as needed for breaking changes or significant feature sets
- **Minor versions** (X.Y.0): Released monthly when new backward-compatible features are added
- **Patch versions** (X.Y.Z): Released as needed for bug fixes, security patches, and improvements

## Release Process

1. Update version numbers in documentation and code
2. Update CHANGELOG.md with release notes
3. Create annotated git tag: `git tag -a vX.Y.Z -m "Release X.Y.Z"`
4. Push tags to remote: `git push origin --tags`
5. Create GitHub release with release notes: `gh release create vX.Y.Z --generate-notes`

## API Stability

**Versions 0.x.x (pre-1.0.0):** API is unstable and may change at any time, even in minor versions.

**Versions 1.x.x and later:** Follow semantic versioning strictly. Breaking changes only occur in major releases and are documented.

## Backwards Compatibility

Within a major version, all public APIs remain stable and backwards compatible with earlier minor and patch versions. Users can safely upgrade within the same major version without code changes.

## Deprecation Policy

Features may be marked as deprecated in a minor or patch release. Deprecated features will continue to work for at least 2 major versions before removal. Users will be notified via:
- Deprecation warnings in logs
- Documentation updates
- Release notes

## Support

- **Current version**: Receives bug fixes and security patches
- **Previous major version**: Security patches only
- **Older versions**: Community support via GitHub Issues

## Version History

- **v0.1.0** - Initial public release
- **v0.2.0** - Enhanced PostgreSQL support
- **v1.0.0** - Production ready (planned)
