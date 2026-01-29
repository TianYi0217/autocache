# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.4] - 2025-10-26

### Fixed
- Real-time streaming with proper HTTP flushing for server-sent events (SSE) (#8)
  - Replaced `io.Copy()` with manual read/write loop with immediate `Flush()` calls
  - Implemented `http.Flusher` interface on `responseWrapper`
  - Use 512-byte buffer optimized for Anthropic SSE events (typically 100-200 bytes)
  - Proper error handling for write vs read errors
  - Graceful degradation when underlying ResponseWriter doesn't support flushing

### Changed
- Streaming responses now deliver content in real-time instead of buffering until completion
- Significantly improved user experience for streaming requests

### Added
- Unit tests for `responseWrapper.Flush()` implementation
- Test coverage for graceful degradation scenarios

**Contributors:** @TianYi0217

## [1.0.3] - 2025-10-25

### Fixed
- System prompt caching for string format (#4, #5)
  - Convert system strings to SystemBlocks when cacheable (>= min tokens)
  - Add proper JSON handling for both string and array formats
  - Add `MarshalJSON` to output SystemBlocks as "system" array
  - Add `UnmarshalJSON` to accept both string and array formats

### Technical Details
- Previously, system prompts using string format were never cached regardless of size
- Cache control can only be applied to blocks format, so strings are now converted when cacheable
- All 81 tests pass with end-to-end API validation

## [1.0.2] - 2025-10-22

### Fixed
- Version information not being set via ldflags (#7)
- System prompt caching for string format (#4, #5)

### Documentation
- Add 'Getting Started with Docker' section to README (#3)

## [1.0.1] - 2025-10-08

### Fixed
- API key header forwarding bug (#2)
  - Fix case-sensitivity bug where duplicate API key headers were sent
  - Remove existing auth headers before adding normalized x-api-key
  - Add debug logging and secure maskAPIKey helper function
  - Proper handling of Authorization, X-Api-Key, and anthropic-api-key headers

## [1.0.0] - 2025-10-08

### Added

#### Core Features
- Intelligent cache-control injection for Anthropic Claude API requests
- Automatic token counting and analysis for optimal cache breakpoint placement
- ROI analytics with detailed cost savings calculations
- Support for multiple cache strategies (conservative, moderate, aggressive)
- Streaming and non-streaming request support
- Response headers with detailed cache metadata (X-Autocache-* headers)

#### Cache Intelligence
- Smart breakpoint placement using ROI scoring algorithm
- Token minimums enforcement (1024 for most models, 2048 for Haiku)
- Automatic TTL assignment (1h for stable content, 5m for dynamic)
- Support for system prompts, tool definitions, and content blocks

#### API Endpoints
- `POST /v1/messages` - Main proxy endpoint with automatic caching
- `GET /health` - Health check endpoint
- `GET /metrics` - Metrics and configuration endpoint
- `GET /savings` - Comprehensive savings analytics and statistics

#### Configuration
- Environment variable configuration support
- Multiple caching strategies with customizable thresholds
- API key handling via headers or environment variables
- Configurable logging (text/JSON, multiple levels)

#### Docker Support
- Multi-stage Dockerfile with optimized build
- Docker Compose configuration with health checks
- Non-root user security
- Resource limits and restart policies

#### Documentation
- Comprehensive README with examples
- Architecture documentation (CLAUDE.md)
- API key handling guide
- n8n integration documentation
- Troubleshooting guides

#### Testing
- Unit tests for all core components
- Real API integration tests
- Test fixtures and utilities
- n8n workflow testing scripts

### Technical Details
- Built with Go 1.23+
- Modular architecture with clean separation of concerns
- Logrus-based structured logging
- Graceful shutdown handling
- HTTPS support for Anthropic API communication

[1.0.4]: https://github.com/montevive/autocache/releases/tag/v1.0.4
[1.0.3]: https://github.com/montevive/autocache/releases/tag/v1.0.3
[1.0.2]: https://github.com/montevive/autocache/releases/tag/v1.0.2
[1.0.1]: https://github.com/montevive/autocache/releases/tag/v1.0.1
[1.0.0]: https://github.com/montevive/autocache/releases/tag/v1.0.0
