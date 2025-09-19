# cat-server

A Go-based HTTP server that provides file system utilities, like an enhanced cat command with REST API capabilities.

## Features

- **Health Check Endpoint**: Monitor server status with `/health`
- **File List Endpoint**: Get directory file listings with `/ls`
- **Flexible Directory Selection**: Configure target directory via command-line flag
- **Hidden File Support**: Includes files starting with dots (hidden files)
- **Multiple Response Formats**: JSON, HTML, and plain text support for health endpoint
- **Structured Logging**: Comprehensive request/response logging with slog

## Quick Start

### Build and Run

```bash
# Build the server
go build -o bin/cat-server src/main.go

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

Get a list of all files in the configured directory (including hidden files).

**Example:**
```bash
curl http://localhost:8080/ls
```

**Response:**
```json
{
  "files": ["README.md", ".gitignore", "main.go", ".hidden"],
  "directory": "./files/",
  "count": 4,
  "generated_at": "2025-09-20T10:00:00Z"
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

# Start server with custom directory
./bin/cat-server -dir ./my-documents
curl http://localhost:8080/ls

# Pretty print JSON output
curl -s http://localhost:8080/ls | jq .

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
go vet ./...      # Static analysis
go fmt ./...      # Code formatting
go test ./...     # Run all tests
go build ./...    # Compilation check
```

### Testing

```bash
# Run all tests
go test ./... -v

# Run specific test suites
go test ./tests/unit/... -v
go test ./tests/integration/... -v
go test ./tests/contract/... -v
go test ./tests/performance/... -v

# Run with coverage
go test ./... -cover
```

### Project Structure

```
├── src/
│   ├── server/          # HTTP server implementation
│   ├── handlers/        # Request handlers (health.go, list.go)
│   ├── services/        # Business logic (directory.go)
│   └── main.go         # Application entry point
├── tests/
│   ├── unit/           # Unit tests
│   ├── integration/    # Integration tests
│   ├── contract/       # API contract tests
│   └── performance/    # Performance/load tests
├── specs/              # Feature specifications
└── bin/                # Compiled binaries
```

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
- `405 Method Not Allowed` - Unsupported HTTP method
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