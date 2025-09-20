# cat-server

A Go-based HTTP server that provides file system utilities, like an enhanced cat command with REST API capabilities.

## Features

- **Health Check Endpoint**: Monitor server status with `/health`
- **File List Endpoint**: Get directory file listings with `/ls`
- **File Content Endpoint**: Read individual file contents with `/cat/{filename}`
- **Clean Architecture**: Domain-driven design with layered architecture
- **Security First**: Path traversal protection and input validation
- **Flexible Directory Selection**: Configure target directory via command-line flag
- **Hidden File Support**: Includes files starting with dots (hidden files)
- **Multiple Response Formats**: JSON, HTML, and plain text support for health endpoint
- **Structured Logging**: Comprehensive request/response logging with slog

## Quick Start

### Build and Run

```bash
# Build the server
go build -o bin/cat-server ./cmd/cat-server/

# Run with default directory (./files/)
./bin/cat-server

# Run with custom directory
./bin/cat-server -dir /path/to/your/directory
```

### API Endpoints

#### Health Check - `GET /health`

Check server health and status.

**Example:**
```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-09-20T10:00:00Z"
}
```

#### File List - `GET /ls`

Get a comprehensive list of all files in the configured directory with detailed metadata.

**Example:**
```bash
curl http://localhost:8080/ls
```

**Response:**
```json
{
  "path": ".",
  "files": [
    {
      "name": "hello.txt",
      "size": 12,
      "sizeHuman": "12 B",
      "modTime": "2025-09-20T19:58:55.580991599+09:00",
      "isDir": false,
      "permissions": "-rw-r--r--",
      "isHidden": false,
      "isExecutable": false,
      "isReadable": true,
      "isWritable": true
    }
  ],
  "totalCount": 1,
  "fileCount": 1,
  "dirCount": 0,
  "totalSize": 12,
  "scannedAt": "2025-09-20T20:52:29.226409+09:00",
  "statistics": {
    "largestFile": { /* file metadata */ },
    "newestFile": { /* file metadata */ },
    "oldestFile": { /* file metadata */ }
  }
}
```

#### File Content - `GET /cat/{filename}`

Read the content of a specific file with metadata and content type detection.

**Example:**
```bash
curl http://localhost:8080/cat/hello.txt
```

**Response:**
```json
{
  "filename": "hello.txt",
  "content": "Hello World\n",
  "size": 12,
  "sizeHuman": "12 B",
  "contentType": "text/plain; charset=utf-8",
  "encoding": "utf-8",
  "isText": true,
  "lineCount": 2,
  "modTime": "2025-09-20T19:58:55.580991599+09:00",
  "readAt": "2025-09-20T20:54:09.932165+09:00",
  "hash": 3639248343
}
```

### Configuration Options

| Flag | Default | Description |
|------|---------|-------------|
| `-dir` | `./files/` | Directory to list files from |

### Examples

```bash
# List files from default directory
curl http://localhost:8080/ls

# Read a specific file
curl http://localhost:8080/cat/hello.txt

# Start server with custom directory
./bin/cat-server -dir ./my-documents
curl http://localhost:8080/ls

# Pretty print JSON output
curl -s http://localhost:8080/ls | jq .
curl -s http://localhost:8080/cat/config.json | jq .

# Check server health
curl http://localhost:8080/health

# Get HTML health response
curl -H "Accept: text/html" http://localhost:8080/health
```

## Development

### Prerequisites

- Go 1.21 or later
- jq (optional, for JSON formatting)

### Quality Gates

All code changes must pass these quality gates:

```bash
go vet ./cmd/cat-server/ ./pkg/... ./internal/...   # Static analysis
go fmt ./cmd/cat-server/ ./pkg/... ./internal/...   # Code formatting
go test ./pkg/... ./internal/...                    # Run all tests
go build ./cmd/cat-server/                          # Compilation check
```

### Testing

```bash
# Run all tests
go test ./pkg/... ./internal/... -v

# Run domain layer tests
go test ./pkg/domain/... -v

# Run infrastructure tests
go test ./pkg/infrastructure/... -v

# Run application tests
go test ./pkg/application/... -v

# Run contract tests (these are designed to fail in TDD approach)
go test ./specs/*/contracts/ -v

# Run with coverage
go test ./pkg/... ./internal/... -cover
```

### Project Structure

The project follows Go standard project layout with Clean Architecture principles:

```
├── cmd/cat-server/              # Application entry point
│   └── main.go                 # Server startup and dependency injection
├── pkg/                        # Public libraries
│   ├── domain/                 # Domain layer (business logic)
│   │   ├── entities/           # Domain entities
│   │   ├── repositories/       # Repository interfaces
│   │   └── valueobjects/       # Value objects
│   ├── application/            # Application layer (use cases)
│   │   └── services/           # Application services
│   └── infrastructure/         # Infrastructure layer
│       ├── filesystem/         # File system implementation
│       ├── http/              # HTTP server and middleware
│       └── logging/           # Logging infrastructure
├── internal/                   # Private application code
│   └── config/                # Configuration management
├── tests/                      # Legacy tests (being refactored)
│   ├── unit/                  # Unit tests
│   ├── integration/           # Integration tests
│   ├── contract/              # API contract tests
│   └── performance/           # Performance/load tests
├── specs/                      # Feature specifications (Specify framework)
└── bin/                        # Compiled binaries
```

### Architecture

The application follows **Clean Architecture** and **Domain Driven Design** principles:

- **Domain Layer**: Contains business entities, value objects, and repository interfaces
- **Application Layer**: Orchestrates domain logic through application services
- **Infrastructure Layer**: Implements external concerns (file system, HTTP, logging)
- **Interfaces Layer**: HTTP handlers and API contracts (integrated in cmd/)

**Key Principles:**
- **Dependency Inversion**: Infrastructure depends on domain abstractions
- **Single Responsibility**: Each layer has a clear purpose
- **Testability**: Domain logic is isolated and easily testable
- **Security**: Input validation and path traversal protection at domain level

## API Specification

### Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "error description",
  "path": "/problematic/path",
  "timestamp": "2025-09-20T10:00:00Z",
  "status_code": 400
}
```

### Status Codes

- `200 OK` - Successful request
- `400 Bad Request` - Invalid directory path or request
- `403 Forbidden` - Permission denied for directory access
- `404 Not Found` - File not found (for `/cat/{filename}`)
- `405 Method Not Allowed` - Unsupported HTTP method
- `413 Payload Too Large` - File size exceeds limit
- `500 Internal Server Error` - Server error

## Security

- Path traversal protection (prevents `../` attacks)
- Null byte injection prevention
- Directory access validation
- File path length limits
- Read permission verification

## Performance

- Target response time: <100ms for directories with <1000 files
- Memory efficient directory reading
- Structured logging for performance monitoring
- Concurrent request support

## License

This project follows Go community best practices and is designed for educational and utility purposes.