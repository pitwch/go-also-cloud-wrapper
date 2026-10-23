# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

# [1.1.0](https://github.com/pitwch/go-also-cloud-wrapper/compare/v1.0.1...v1.1.0) (2026-10-23)


### Features

* harden HTTP client and support custom HTTPClient ([5e94069](https://github.com/pitwch/go-also-cloud-wrapper/commit/5e9406933ef06873bbded7946b0b29c37aa53512))

## [1.0.1](https://github.com/pitwch/go-also-cloud-wrapper/compare/v1.0.0...v1.0.1) (2025-10-02)


### Bug Fixes

* improve error handling and resource cleanup in HTTP client ([a427f40](https://github.com/pitwch/go-also-cloud-wrapper/commit/a427f403ceedba3e290845b0641ed41c559cde02))

# 1.0.0 (2025-10-02)


### Features

* add testing instructions and update Swagger API URL, ci fixes ([d65f601](https://github.com/pitwch/go-also-cloud-wrapper/commit/d65f601b69473e9edc77888204b4f881e0bb8e65))
* Migrate to github workflow, fix and refactor ([95f0eaa](https://github.com/pitwch/go-also-cloud-wrapper/commit/95f0eaa0b195d5f28bf9106501558489baea4596))

## [Unreleased]

### Added

- New HTTP methods: `Get()`, `Put()`, `Delete()`, `Patch()` for full REST API support
- Timeout configuration now properly applied to HTTP client
- Better logging with prefixed log messages for easier debugging
- Comprehensive error handling with nil pointer checks
- Context-aware request handling
- Better test coverage with timeout and error tests

### Fixed
- **CRITICAL**: Fixed potential nil pointer dereference in `request()` method when accessing session token
- **CRITICAL**: Fixed resource leaks - HTTP response bodies now properly closed with defer
- Fixed error handling in `createNewToken()` - no longer crashes on nil response
- Fixed `Validate()` method to properly handle errors and close response body
- Fixed `errorFormatterPx()` to handle nil readers and unparseable XML gracefully
- Fixed inefficient boolean returns in `Validate()` method
- Error messages now properly formatted with structured information

### Changed
- HTTP client initialization improved with proper timeout configuration
- `strings.Trim()` now used instead of `strings.Replace()` for token cleanup (more efficient)
- Error messages now use structured format with proper error wrapping
- Log messages now include `[go-also-cloud-wrapper]` prefix for clarity
- Improved error handling throughout with proper error wrapping using `%w`
- Boolean comparisons simplified (removed redundant `== true` checks)
- Test improvements: added defer for resource cleanup, better skip logic

### Optimized
- Reduced unnecessary buffer allocations
- Better memory management with proper defer patterns
- More efficient string operations
- Removed redundant error checks

## [1.0.0] - Initial Release

### Added
- Initial ALSO Cloud Marketplace API wrapper
- Basic authentication with session token management
- POST request support
- Session validation
- Error handling with XML parsing
- Support for multiple marketplace regions
- Configurable options (APIPrefix, LoginEndpoint, UserAgent, Batchsize, Log)
