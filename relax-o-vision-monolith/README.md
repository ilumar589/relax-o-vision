# Relax-O-Vision

Welcome to Relax-O-Vision - a comprehensive football data analytics platform with AI-powered predictions! 🎉

## Quick Links

- [Getting Started Guide](GETTING_STARTED.md) - Complete setup instructions for new developers
- [Football Predictions](FOOTBALL_PREDICTIONS.md) - AI prediction features
- [Implementation Summary](IMPLEMENTATION_SUMMARY.md) - Technical architecture details

This README file contains all the necessary information about:

- [Project overview](#project-overview)
- [Folders structure](#folders-structure)
- [Starting your project](#starting-your-project)
- [Developing your project](#developing-your-project)
- [Testing](#testing)
- [Deploying your project](#deploying-your-project)

And some words [about the Gowebly CLI](#about-the-gowebly-cli).

## Project overview

Backend:

- Module name in the go.mod file: `github.com/edd/relaxovisionmonolith`
- Go web framework/router: `Fiber`
- Server port: `7000`

Frontend:

- Package name in the package.json file: `relaxovisionmonolith`
- Reactivity library: `htmx with Alpine.js`
- CSS framework: `Tailwind CSS with Flowbite components`

Tools:

- Air tool to live-reloading: ✓
- Bun as a frontend runtime: ✓
- Templ to generate HTML: ✓
- Config for golangci-lint: ✓

## Folders structure

```console
.
├── assets
│   ├── scripts.js
│   └── styles.scss
├── static
│   ├── images
│   │   └── gowebly.svg
│   ├── apple-touch-icon.png
│   ├── favicon.ico
│   ├── favicon.png
│   ├── favicon.svg
│   ├── manifest-desktop-screenshot.jpeg
│   ├── manifest-mobile-screenshot.jpeg
│   ├── manifest-touch-icon.svg
│   └── manifest.webmanifest
├── templates
│   ├── pages
│   │   └── index.templ
│   └── main.templ
├── .gitignore
├── .dockerignore
├── .prettierignore
├── .air.toml
├── golangci.yml
├── Dockerfile
├── docker-compose.yml
├── prettier.config.js
├── package.json
├── go.mod
├── go.sum
├── handlers.go
├── server.go
├── main.go
└── README.md
```

## Starting your project

> ❗️ Please make sure that you have installed the executable files for all the necessary tools before starting your project. Exactly:
>
> - `Air`: [https://github.com/air-verse/air](https://github.com/air-verse/air)
> - `Bun`: [https://github.com/oven-sh/bun](https://github.com/oven-sh/bun)
> - `Templ`: [https://github.com/a-h/templ](https://github.com/a-h/templ)
> - `golangci-lint`: [https://github.com/golangci/golangci-lint](https://github.com/golangci/golangci-lint)

To start your project, run the **Gowebly** CLI command in your terminal:

```console
gowebly run
```

## Developing your project

The backend part is located in the `*.go` files in your project folder.

The `./templates` folder contains Templ templates that you can use in your frontend part. Also, the `./assets` folder contains the `styles.scss` (main styles) and `scripts.js` (main scripts) files.

The `./static` folder contains all the static files: icons, images, PWA (Progressive Web App) manifest and other builded/minified assets.

## Testing

This project includes a comprehensive test suite for the football data caching system, scheduler, and cache manager components.

### Running Tests

#### Run All Tests

```bash
GOEXPERIMENT=jsonv2 go test ./...
```

> **Note**: This project uses Go 1.25's experimental JSON v2 implementation. The `GOEXPERIMENT=jsonv2` environment variable must be set when running tests and building the project.

#### Run Tests with Verbose Output

```bash
GOEXPERIMENT=jsonv2 go test -v ./...
```

#### Run Tests with Coverage

```bash
GOEXPERIMENT=jsonv2 go test -cover ./...
```

#### Generate Coverage Report

```bash
GOEXPERIMENT=jsonv2 go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### Run Specific Package Tests

```bash
# Football data tests
GOEXPERIMENT=jsonv2 go test -v ./footballdata/...

# Cache tests
GOEXPERIMENT=jsonv2 go test -v ./cache/...
```

#### Run Integration Tests

```bash
# Requires Docker for testcontainers
GOEXPERIMENT=jsonv2 go test -v -tags=integration ./...
```

### Test Structure

| Package | File | Description |
|---------|------|-------------|
| `footballdata` | `cache_manager_test.go` | Tests for cache coordination between Redis and PostgreSQL |
| `footballdata` | `scheduler_test.go` | Tests for background sync scheduler and refresh logic |
| `footballdata` | `client_test.go` | Tests for football-data.org API client |
| `footballdata` | `repository_test.go` | Tests for PostgreSQL data access |
| `footballdata` | `service_test.go` | Tests for business logic layer |
| `footballdata` | `mocks_test.go` | Mock implementations for testing |
| `footballdata` | `testutil_test.go` | Test fixtures and helper functions |
| `cache` | `redis_test.go` | Tests for Redis cache implementation |
| `cache` | `cache_test.go` | Tests for cache interface and memory cache |
| `testutil` | `docker.go` | Utilities for integration tests with Docker containers |

### Test Categories

#### Unit Tests

Fast, isolated tests with mocked dependencies. Run without external services.

```bash
GOEXPERIMENT=jsonv2 go test -short ./...
```

Unit tests include:
- Cache manager operations (Get/Set/Delete)
- Scheduler logic and lifecycle
- Client rate limiting
- Mock implementations
- Concurrent access patterns

#### Integration Tests

Tests with real PostgreSQL and Redis via testcontainers. Requires Docker.

```bash
GOEXPERIMENT=jsonv2 go test -tags=integration ./...
```

Integration tests include:
- Database CRUD operations
- Cache TTL and expiration
- JSONB column handling
- Upsert (ON CONFLICT) behavior
- End-to-end data flow

### Key Test Scenarios

#### Cache Manager Tests
- ✅ Cache hit returns cached data
- ✅ Cache miss returns nil gracefully
- ✅ Expired cache triggers refresh
- ✅ Redis failure is handled gracefully
- ✅ Concurrent access is thread-safe
- ✅ Data hash computation is consistent

#### Scheduler Tests
- ✅ needsRefresh() logic works correctly
- ✅ Scheduler start/stop lifecycle functions
- ✅ Context cancellation stops scheduler
- ✅ Competition codes are processed correctly
- ✅ Rate limiting prevents API overload

#### API Client Tests
- ✅ Correct headers are set (X-Auth-Token, Accept)
- ✅ Rate limiting (10 req/min) prevents excessive calls
- ✅ HTTP errors are handled gracefully
- ✅ Concurrent requests are thread-safe
- ✅ Mock HTTP servers work correctly

#### Repository Tests
- ✅ Competitions are saved with JSONB data
- ✅ cached_at timestamp is updated on save
- ✅ Upsert works correctly (INSERT ON CONFLICT)
- ✅ Queries return correct data
- ✅ JSONB columns serialize/deserialize properly

#### Cache Tests
- ✅ Memory cache Get/Set/Delete operations
- ✅ Memory cache Clear removes all items
- ✅ Memory cache handles concurrent access
- ✅ Cache factory creates correct implementations
- ✅ Default to memory cache when type unknown

### Writing New Tests

Follow these conventions:

1. **Use table-driven tests** for comprehensive coverage:

```go
func TestNeedsRefresh(t *testing.T) {
    tests := []struct {
        name     string
        cachedAt time.Time
        expected bool
    }{
        {
            name:     "fresh data (1 day old)",
            cachedAt: time.Now().Add(-24 * time.Hour),
            expected: false,
        },
        {
            name:     "stale data (31 days old)",
            cachedAt: time.Now().Add(-31 * 24 * time.Hour),
            expected: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

2. **Name test functions** as `Test<Function>_<Scenario>`
3. **Use `t.Parallel()`** for independent tests
4. **Clean up resources** in `t.Cleanup()` or defer
5. **Use `t.Skip()`** for integration tests that require external services

### Mocking

Mock implementations are available in `*_test.go` files:
- `MockCache` - For testing cache operations without Redis
- `MockRepository` - For testing without database
- `MockHTTPClient` - For testing API client without actual HTTP calls

Example usage:

```go
func TestCacheManager_Get_CacheHit(t *testing.T) {
    mockCache := NewMockCache()
    cm := &CacheManager{redis: mockCache}
    
    // Set up test data
    mockCache.Set(ctx, "key", []byte("value"), 1*time.Hour)
    
    // Test cache hit
    data, err := cm.Get(ctx, "key")
    // Assertions...
}
```

### Benchmark Tests

Benchmark tests are included for performance-critical operations:

```bash
GOEXPERIMENT=jsonv2 go test -bench=. ./...
```

Available benchmarks:
- `BenchmarkCacheManager_Get`
- `BenchmarkCacheManager_Set`
- `BenchmarkClient_RateLimitCheck`
- `BenchmarkScheduler_needsRefresh`
- `BenchmarkMemoryCache_Get`
- `BenchmarkMemoryCache_Set`
- `BenchmarkMemoryCache_ConcurrentGet`

### CI/CD

Tests run automatically on:
- Pull request creation
- Push to main branch

The CI workflow uses `GOEXPERIMENT=jsonv2` for all test runs. Coverage reports are generated and can be viewed in the GitHub Actions artifacts.

### Edge Cases Tested

- Empty API responses
- Malformed JSON responses
- Network timeouts
- Database connection failures
- Redis connection failures  
- Concurrent sync operations
- Large datasets
- Unicode in team/competition names
- Null/missing fields in API responses

### Test Coverage

Current test coverage:
- `footballdata` package: Unit tests with mocks
- `cache` package: Unit tests with memory cache, integration tests for Redis

To improve coverage, run:

```bash
GOEXPERIMENT=jsonv2 go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Deploying your project

All deploy settings are located in the `Dockerfile` and `docker-compose.yml` files in your project folder.

To deploy your project to a remote server, follow these steps:

1. Go to your hosting/cloud provider and create a new VDS/VPS.
2. Update all OS packages on the server and install Docker, Docker Compose and Git packages.
3. Use `git clone` command to clone the repository with your project to the server and navigate to its folder.
4. Run the `docker-compose up` command to start your project on your server.

> ❗️ Don't forget to generate Go files from `*.templ` templates before run the `docker-compose up` command.

## About the Gowebly CLI

The [**Gowebly**](https://github.com/gowebly/gowebly) CLI is a next-generation CLI tool that makes it easy to create amazing web applications with **Go** on the backend, using **htmx**, **hyperscript** or **Alpine.js**, and the most popular **CSS frameworks** on the frontend.

It's highly recommended to start exploring the Gowebly CLI with short articles "[**What is Gowebly CLI?**](https://gowebly.org/getting-started)" and "[**How does it work?**](https://gowebly.org/getting-started/how-does-it-work)" to understand the basic principle and the main components built into the **Gowebly** CLI.

<a href="https://gowebly.org/" target="_blank"><img height="112px" alt="another awesome project built with the Gowebly CLI" src="https://raw.githubusercontent.com/gowebly/.github/main/images/gowebly-new-project-banner.svg"/></a>
